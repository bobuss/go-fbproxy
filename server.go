package main

import (
	"fmt"
	"html"
	"log"
  "sync"
	"net/http"
	"github.com/gorilla/mux"
	fb "github.com/huandu/facebook"
  //"github.com/garyburd/redigo/redis"
)

// const (
//     ADDRESS = "127.0.0.1:6379"
// )

var (
  mutex = &sync.Mutex{}
  memory = make(map[string]string)
  //conn, err = redis.Dial("tcp", ADDRESS)
)

func main() {
  // if err != nil {
  //   log.Fatal(err)
  // }

  rtr := mux.NewRouter()
  rtr.HandleFunc("/user/{fbuid:[0-9]+}/profile", profile).Methods("GET")

  http.Handle("/", rtr)

  log.Println("Listening...")
  http.ListenAndServe(":3000", nil)
}

func profile(w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  fbuid := params["fbuid"]

  messages := make(chan string)

  go func(fbuid string) {

    var key string
    key = "fbproxy:" + fbuid + ":profile"

    // n, _ := conn.Do("GET", key)

    mutex.Lock()
    out, ok := memory[key]
    if !ok {

      log.Println("Calling Facebook")
      res, _ := fb.Get("/" + fbuid, fb.Params{
        "fields": "username",
      })
      var username string
      res.DecodeField("username", &username)
      memory[key] = username
      messages <- username

    } else {
      messages <- out
    }
    mutex.Unlock()

  }(fbuid)

  fmt.Fprintf(w, "Hello, %q", html.EscapeString(<- messages))
}