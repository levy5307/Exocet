package main 

import (
    "fmt"
    "github.com/garyburd/redigo/redis"
    "regexp"
    "github.com/AlexStocks/goext/time" // gxtime
    "time"
)

var wh *gxtime.Wheel
var ch chan string 

// get sentinel info once
func getSentinelInfo() (str string)  {

    // connect redis
    conn, err := redis.Dial("tcp", "192.168.160.137:30000")
    if err != nil {
        fmt.Println("connect error: ", err)
        return
    }
    defer conn.Close()

    // get result string
    all_str, err := redis.String(conn.Do("info", "sentinel"))
    if err != nil {
        fmt.Println("Do func error: ", err)
        return
    }

   // parse name, using regular expression
   name_reg := regexp.MustCompile(`(master[0-9]{1,}):name=([^0-9]{1,})[0-9]`)

   // parse address:port, using regular expression
   address_reg := regexp.MustCompile(`address=(([0-9]{1,}.){3})([0-9]{1,}):([0-9]{1,})`)

   str = fmt.Sprintln(name_reg.FindAllString(all_str, -1)) + fmt.Sprintln(address_reg.FindAllString(all_str, -1))

   return
}

// get sentinel info periodically
func GetSentinelInfo(wh *gxtime.Wheel, span time.Duration) {
    var (
        infoMap map[int] string
        version int
        infoSen string
    )
    version = 1
    infoMap = make(map[int] string)

    for {
        select {
        case <- wh.After(span):
            infoSen = getSentinelInfo()
        case infoChan := <-ch:
            infoSen = getSentinelInfo() + infoChan
        }
        fmt.Println("version: ", version, infoSen)
        infoMap[version] = infoSen
        version++
    }
}

func SubscribeODown() {

    // connect redis
    conn, err := redis.Dial("tcp", "192.168.160.137:30000")
    if err != nil {
        fmt.Println("connect error: ", err)
        return
    }
    defer conn.Close()

    psc := redis.PubSubConn{conn}
    psc.Subscribe("+odown")
    for {
        switch v := psc.Receive().(type) {
        case redis.Message:
            fmt.Printf("%s: message: %s", v.Channel, v.Data)
        case redis.Subscription:
            ch <- fmt.Sprintf("%s: %s %d", v.Channel, v.Kind, v.Count)
        case error:
            fmt.Println(v)
            return
        }
    }
}

func main() {
    var (
        wh *gxtime.Wheel
    )

    ch = make(chan string)
    defer close(ch)

    fmt.Println("start listen...")

    // create a new wheel
    wh = gxtime.NewWheel(gxtime.TimeMillisecondDuration(10000), 20)

    go SubscribeODown()

    // get sentinel info periodically with timewheel
    GetSentinelInfo(wh, gxtime.TimeMillisecondDuration(30000))
}
