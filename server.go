package main

import (
	"fmt"
	"html"
	"log"
  "sync"
	"net/http"
	"github.com/gorilla/mux"
	fb "github.com/huandu/facebook"
  "github.com/garyburd/redigo/redis"
  flag "github.com/ogier/pflag"
  storage "./storage"
)


var (
  redisAddress   = flag.String("redis-address", "127.0.0.1:6379", "Address to the Redis server")
  maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
  ttl = flag.Int("ttl", 300, "Time-to-live of the backend")
  mutex = &sync.Mutex{}
)


func SetupMemory() storage.Storage {
  m := make(map[string]string)
  return storage.Memory{m}
}


func SetupRedis() storage.Storage {
  p := redis.NewPool(func() (redis.Conn, error) {
    c, err := redis.Dial("tcp", *redisAddress)

    if err != nil {
      panic(err)
    }

    return c, err
  }, *maxConnections)

  store := storage.Redis{p}
  return store
}


func profileHander(w http.ResponseWriter, r *http.Request, store storage.Storage) {
  params := mux.Vars(r)
  fbuid := params["fbuid"]

  messages := make(chan string)

  go func(fbuid string) {

    var key string
    key = "fbproxy:" + fbuid + ":profile"

    n := store.Get(key)

    if n == "" {

      mutex.Lock()

      // check a second time, if a value was set by another goroutine
      n := store.Get(key)

      if n == "" {

        log.Println("Calling Facebook")
        res, _ := fb.Get("/" + fbuid, fb.Params{
          "fields": "username",
        })
        var username string
        res.DecodeField("username", &username)
        store.Set(key, username, 300)
        messages <- username

      } else {
        username, _ := redis.String(n, nil)
        messages <- username
      }
      mutex.Unlock()
    } else {
      username, _ := redis.String(n, nil)
      messages <- username
    }

  }(fbuid)

  fmt.Fprintf(w, "Hello, %q", html.EscapeString(<- messages))
}


func main() {

  flag.Parse()

  //store := SetupRedis()
  store := SetupMemory()
  //defer redisPool.Close()

  router := mux.NewRouter()

  router.HandleFunc("/user/{fbuid:[0-9]+}/profile", func (w http.ResponseWriter, r *http.Request) {
    profileHander(w, r, store)
  }).Methods("GET")

  router.HandleFunc("/ping/", func (w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "pong")
  }).Methods("GET")

  http.Handle("/", router)

  log.Println("Listening...")
  http.ListenAndServe(":3000", nil)
}