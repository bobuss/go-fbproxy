package main

import (
	"fmt"
	"html"
	"log"
  "sync"
	"net/http"
	"github.com/gorilla/mux"
	fb "github.com/huandu/facebook"
)


var (
  mutex = &sync.Mutex{}
  memory = make(map[string]string)
)

func main() {
  rtr := mux.NewRouter()
  rtr.HandleFunc("/user/{fbuid:[0-9]+}/profile", profile).Methods("GET")

  http.Handle("/", rtr)

  log.Println("Listening...")
  http.ListenAndServe(":3000", nil)
}

func profile(w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  fbuid := params["fbuid"]

  out, ok := memory[fbuid]

  if !ok {
    mutex.Lock()
    out, ok := memory[fbuid]
    if !ok {
      fmt.Println("Calling Facebook")
      res, _ := fb.Get("/" + fbuid, fb.Params{
        "fields": "username",
      })
      out = res["username"].(string)
      memory[fbuid] = out
    }
    mutex.Unlock()
  }

  fmt.Fprintf(w, "Hello, %q", html.EscapeString(out))
}