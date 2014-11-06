package main

import (
	"fmt"
	"log"
  "sync"
  "io"
  "bytes"
	"net/http"
	"github.com/gorilla/mux"
  flag "github.com/ogier/pflag"
  storage "./storage"
)

const GRAPH_API = "https://graph.facebook.com"

var (
  redisAddress = flag.String("redis-address", "127.0.0.1:6379", "Address to the Redis server")
  ttl = flag.Int("ttl", 300, "Time-to-live of the backend")
  mutex = &sync.Mutex{}
)


func SetupMemory() storage.Storage {
  m := make(map[string]string)
  return storage.Memory{m}
}


func SetupRedis() storage.Storage {
  p := storage.NewPool(*redisAddress)
  store := storage.Redis{p}
  return store
}


func profileHander(w http.ResponseWriter, r *http.Request, store storage.Storage) {
  params := mux.Vars(r)
  fbuid := params["fbuid"]
  request_path := r.URL.Query().Get("fields")

  messages := make(chan []byte)

  go func(fbuid, request_path string) {

    var key string
    key = "fbproxy:" + fbuid + "/" + request_path

    read := store.Get(key)

    if read == "" {

      mutex.Lock()

      // check a second time, if a value was set by another goroutine
      read := store.Get(key)

      if read == "" {

        client := &http.Client{}

        fb_path := GRAPH_API + "/" + fbuid + "?fields=" + request_path

        log.Println("Calling Facebook : " + fb_path)

        req, err := http.NewRequest(r.Method, fb_path, nil)
        req.Header = r.Header

        response, err := client.Do(req)

        defer response.Body.Close()

        buf := &bytes.Buffer{}
        _, err = io.Copy(buf, response.Body)

        if err != nil {
          panic(err)
        }

        ret := buf.Bytes()
        messages <- ret
        store.Set(key, string(ret), *ttl)

      } else {
        messages <- []byte(read)
      }
      mutex.Unlock()
    } else {
      messages <- []byte(read)
    }

  }(fbuid, request_path)

  w.Header().Set("Content-Type", "application/json")
  w.Write(<- messages)
}


func main() {

  flag.Parse()

  store := SetupRedis()
  //store := SetupMemory()

  router := mux.NewRouter()

  router.HandleFunc("/proxy/{fbuid:[0-9]+}", func (w http.ResponseWriter, r *http.Request) {
    profileHander(w, r, store)
  }).Methods("GET")

  router.HandleFunc("/proxy/{fbuid:[0-9]+}/{request_path:.*}", func (w http.ResponseWriter, r *http.Request) {
    profileHander(w, r, store)
  }).Methods("GET")

  router.HandleFunc("/ping/", func (w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "pong")
  }).Methods("GET")

  http.Handle("/", router)

  log.Println("Listening...")
  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}