package asd

import (
  "fmt"
  "sync"
  "testing"
)

var m = sync.Map{}

func TestName(t *testing.T) {
  const key = "t"
  for i := 0; i < 100000; i++ {
    go func(i int) {
      ob, ok := m.LoadOrStore(key, &i)
      if ok {
        fmt.Println("ok = ", i)
      }
      fmt.Printf("%v \n", ob)
    }(i)
  }
}
