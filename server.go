package main

import (
	"fmt"
  "flag"
	"html"
	"log"
  "sync"
	"net/http"
	"github.com/gorilla/mux"
	fb "github.com/huandu/facebook"
  "github.com/garyburd/redigo/redis"
)

const (
    ADDRESS = "127.0.0.1:6379"
)

var (
  redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
  maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
  mutex = &sync.Mutex{}
)

func SetupRedis() *redis.Pool {
  return redis.NewPool(func() (redis.Conn, error) {
    c, err := redis.Dial("tcp", *redisAddress)

    if err != nil {
      panic(err)
    }

    return c, err
  }, *maxConnections)
}


func main() {

  flag.Parse()

  redisPool := SetupRedis()
  defer redisPool.Close()

  rtr := mux.NewRouter()
  rtr.HandleFunc("/user/{fbuid:[0-9]+}/profile", func (w http.ResponseWriter, r *http.Request) {

    profile(w, r, redisPool)

  }).Methods("GET")

  http.Handle("/", rtr)

  log.Println("Listening...")
  http.ListenAndServe(":3000", nil)
}

func profile(w http.ResponseWriter, r *http.Request, pool *redis.Pool) {
  params := mux.Vars(r)
  fbuid := params["fbuid"]

  messages := make(chan string)

  go func(fbuid string) {

    var key string
    key = "fbproxy:" + fbuid + ":profile"

    conn := pool.Get()
    defer conn.Close()

    n, err := conn.Do("GET", key)

    if err != nil {
      log.Println(err)
    }

    if n == nil {

      mutex.Lock()

      // check a second time, if a value was set by another goroutine
      n, _ := conn.Do("GET", key)

      if n == nil {

        log.Println("Calling Facebook")
        res, _ := fb.Get("/" + fbuid, fb.Params{
          "fields": "username",
        })
        var username string
        res.DecodeField("username", &username)
        conn.Do("SET", key, username)
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