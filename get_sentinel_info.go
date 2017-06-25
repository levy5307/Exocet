package main 

import (
    "fmt"
    "github.com/garyburd/redigo/redis"
    "regexp"
    "github.com/AlexStocks/goext/time" // gxtime
    "time"
)

var wh *gxtime.Wheel

// get sentinel info once
func getSentinelInfo() {

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
   fmt.Println(name_reg.FindAllString(all_str, -1))

   // parse address:port, using regular expression
   address_reg := regexp.MustCompile(`address=(([0-9]{1,}.){3})([0-9]{1,}):([0-9]{1,})`)
   fmt.Println(address_reg.FindAllString(all_str, -1))

}

// get sentinel info periodically
func GetSentinelInfo(wh *gxtime.Wheel, span time.Duration) {

    for {
        select {
        case <-wh.After(span):
            getSentinelInfo()
        }
    }

}

func main() {
   
    var (
        wh *gxtime.Wheel
    )

    // create a new wheel
    wh = gxtime.NewWheel(gxtime.TimeMillisecondDuration(100), 20)

    // get sentinel info periodically with timewheel
    GetSentinelInfo(wh, gxtime.TimeMillisecondDuration(300))

}
