package asd

import (
  "fmt"
  "sync"
  "testing"
  "time"
)

var m = sync.Map{}

func TestName(t *testing.T) {
  //n := 0
  //OnceInMem("xx", time.Second*100, func() (interface{}, error) {
  //  return 1, nil
  //}, &n)
  //
  //fmt.Println("n", n)
  //
  //return

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

var desc = "财联社10月25日电，国家数据局党组书记、局长刘烈宏10月25日在2024年“数据要素×”大赛全国总决赛颁奖仪式上表示，一年来，已经出台公共数据开发利用、数字经济高质量发展、城市全域数字化转型、数字经济促进共同富裕、“数据要素×”三年行动计划、全国一体化算力网、国家数据标准体系建设指南等方面的重要政策文件8份。刘烈宏提到，后续拟出台企业数据开发利用、数据产业高质量发展、公共数据资源登记、公共数据授权运营、数据空间、数据流通安全治理等方面的7份政策文件，还将推动建立健全数据产权制度，完善数据流通交易规则，提升数据安全治理水平，促进数据“供得出、流得动、用得好、保安全"

type Obj struct {
  Id   int
  Name string
  Age  int
  Desc string
}

func TestMap(t *testing.T) {

  fun := func() {
    var _map = sync.Map{}
    for i := 0; i < 20000; i++ {
      _map.Store(i, Obj{
        Id:   i,
        Name: "",
        Age:  0,
        Desc: desc,
      })
    }

    begin := time.Now().UnixNano()
    begin = begin / 1e6
    _map.Range(func(key, value interface{}) bool {
      _map.Delete(key)
      return true
    })

    dur := time.Now().UnixNano()
    dur = dur / 1e6
    fmt.Println("=====cost: ", dur-begin)
  }

  for i := 0; i < 5000; i++ {
    go fun()
  }

  time.Sleep(5 * time.Minute)

}
