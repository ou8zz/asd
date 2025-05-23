package asd

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/gomodule/redigo/redis"
  "math"
  "reflect"
  "sync"
  "time"
)

type onceVo struct {
  Once      *sync.Once
  ExpiresAt time.Time
  Error     error
  Data      interface{}
}

var onceMap sync.Map

func init() {
  go func() {
    for true {
      clearExpiresKey()
      time.Sleep(1 * time.Second)
    }
  }()
}

func clearExpiresKey() {
  i := 0
  startTime := time.Now()
  onceMap.Range(func(key, value interface{}) bool {
    v := value.(*onceVo)
    if v.ExpiresAt.Add(1 * time.Second).Before(time.Now()) {
      onceMap.Delete(key)
    }
    i++
    return true
  })
  if n := time.Now().Sub(startTime).Milliseconds(); n > 100 {
    fmt.Printf("clearExpiresKey expires key cost: %d \n", n)
  }
  if i > 10000 {
    fmt.Printf("clearExpiresKey keys max10000 length: %d \n", i)
  }
}

func setV(source, dst interface{}) error {
  // ValueOf to enter reflect-land
  dstPtrValue := reflect.ValueOf(dst)
  if dstPtrValue.Kind() != reflect.Ptr {
    return errors.New("destination must be kind of ptr")
  }
  if dstPtrValue.IsNil() {
    return errors.New("destination cannot be nil")
  }
  //dstType := dstPtrType.Elem()
  // the *dst in *dst = zero
  dstValue := reflect.Indirect(dstPtrValue)
  // the = in *dst = 0
  dstValue.Set(reflect.ValueOf(source))
  return nil
}

func unmarshalFromRedis(conn redis.Conn, key string, dst interface{}) error {
  bytes, err := redis.Bytes(conn.Do("GET", key))
  if err != nil {
    return err
  }
  return json.Unmarshal(bytes, dst)
}

func loadOnce(key string, duration time.Duration) *onceVo {
  onceObj, ok := onceMap.LoadOrStore(key, &onceVo{
    Once:      &sync.Once{},
    ExpiresAt: time.Now().Add(duration),
  })
  if ok {
    //once := onceObj.(*onceVo)
    //if once.ExpiresAt.Before(time.Now()) {
    //  onceMap.Delete(key)
    //  return once
    //}
  }
  return onceObj.(*onceVo)
}

func RemoveOnceMap(key string) {
  onceMap.Delete(key)
}

// 使用场景: 执行函数中只有 err==nil 且 data!=nil 时才进行缓存
func OnceInMem(key string, duration time.Duration, fallback func() (interface{}, error), dst interface{}) error {
  newOnce := loadOnce(key, duration)

  newOnce.Once.Do(func() {
    newOnce.Data, newOnce.Error = fallback()
    if newOnce.Error != nil {
      onceMap.Delete(key)
    }
  })
  if newOnce.Error != nil {
    return newOnce.Error
  }
  if newOnce.Data != nil {
    setV(newOnce.Data, dst)
  } else {
    onceMap.Delete(key)
    fmt.Println("OnceInMem data is nil", key)
  }
  return nil
}

// 使用场景: 忽略错误,默认缓存所有
func OnceInCache(key string, duration time.Duration, fallback func() (interface{}, error), dst interface{}) error {
  newOnce := loadOnce(key, duration)
  newOnce.Once.Do(func() {
    newOnce.Data, newOnce.Error = fallback()
  })
  if newOnce.Data != nil {
    setV(newOnce.Data, dst)
  }
  if newOnce.Error != nil {
    return newOnce.Error
  }
  return nil
}

func OnceInRedis(key string, duration time.Duration, fallback func() (interface{}, error), dst interface{}) error {
  newOnce := loadOnce(key, duration)

  newOnce.Once.Do(func() {
    newOnce.Data, newOnce.Error = fallback()
    if newOnce.Error != nil {
      fmt.Errorf("get data error: %s", newOnce.Error.Error())
      //conn.Do("DEL", key) 不用删除，等待自动过期
    } else {
      //if newOnce.Error = setV(newOnce.Data, dst); newOnce.Error != nil {
      //  return
      //}
      var bytes []byte
      bytes, newOnce.Error = json.Marshal(newOnce.Data)
      if newOnce.Error != nil {
        return
      }
      conn := redisPool.Get()
      defer conn.Close()
      _, newOnce.Error = conn.Do("SET", key, bytes)
      if newOnce.Error != nil {
        return
      }

      var expireTime = math.Max(math.Ceil(duration.Seconds()*2), 1)
      _, newOnce.Error = conn.Do("EXPIRE", key, expireTime)
    }
  })
  if newOnce.Error != nil {
    onceMap.Delete(key)
    return newOnce.Error
  }
  if newOnce.Data != nil {
    conn := redisPool.Get()
    defer conn.Close()
    err := unmarshalFromRedis(conn, key, dst)
    if err != nil {
      onceMap.Delete(key)
      setV(newOnce.Data, dst)
    }
  }
  return nil
}
