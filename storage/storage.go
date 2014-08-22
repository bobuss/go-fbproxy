package storage

import (
  "log"
  "time"
  "github.com/garyburd/redigo/redis"
)

type Storage interface {
   Get(key string) string
   Set(key string, value string, ttl int)
}

/*
Redis Storage
*/
type Redis struct {
  Pool *redis.Pool
}

func NewPool(server string) *redis.Pool {
  return &redis.Pool{
    MaxIdle: 3,
    IdleTimeout: 240 * time.Second,
    Dial: func () (redis.Conn, error) {
      c, err := redis.Dial("tcp", server)
      if err != nil {
        return nil, err
      }
      return c, err
    },
    TestOnBorrow: func(c redis.Conn, t time.Time) error {
      _, err := c.Do("PING")
      return err
    },
}
}

func (r Redis) Get(key string) string {
    conn := r.Pool.Get()
    defer conn.Close()

    n, err := conn.Do("GET", key)

    if err != nil {
      log.Println(err)
      panic(err)
    }
    res, _ := redis.String(n, nil)
    return res
}

func (r Redis) Set(key string, value string, ttl int) {
    conn := r.Pool.Get()
    defer conn.Close()

    _, err := conn.Do("SET", key, value, "EX", ttl)

    if err != nil {
      log.Println(err)
      panic(err)
    }
}


/*
Memory Storage
*/
type Memory struct {
    Dict map[string]string
}

func (m Memory) Get(key string) string {
    return m.Dict[key]
}

func (m Memory) Set(key string, value string, ttl int) {
    m.Dict[key] = value
}