package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

import (
	"github.com/AlexStocks/goext/time" // gxtime
	Log "github.com/AlexStocks/log4go"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

var (
	ch      chan string
	version int32
	infoMap map[int][]Meta
)

const (
	SentinelAddr = "192.168.1.105:40000"
	DBAddr = "192.168.1.105:20000"
)

type NetAddress struct {
	IP   string
	Port int
}

type Meta struct {
	Name string

	Master NetAddress
	Slaves []NetAddress
}

func getSlavesInfo(masterAddr NetAddress) ([]NetAddress, error) {
	var (
		master string
		slaves []NetAddress
		slave  NetAddress
	)

	master = fmt.Sprintf("%s:%s", masterAddr.IP, strconv.Itoa(masterAddr.Port))

	// connect redis
	conn, err := redis.Dial("tcp", master)
	if err != nil {
		return slaves, errors.Wrapf(err, fmt.Sprintf("fail to connect master %s", master))
	}
	defer conn.Close()

	// get result string
	allStr, err := redis.String(conn.Do("info", "Replication"))
	if err != nil {
		return slaves, errors.Wrapf(err, fmt.Sprintf("fail to exec command info:Replication"))
	}

	// parse slave ip, using regular expression
	ipReg := regexp.MustCompile(`ip=(([0-9]{1,}.){3}[0-9])`)
	ips := ipReg.FindAllString(allStr, -1)

	// parse slave port
	portReg := regexp.MustCompile(`port=[0-9]{1,}`)
	ports := portReg.FindAllString(allStr, -1)

	// if ip count != port count, return error
	if len(ips) != len(ports) {
		return slaves, errors.Wrapf(err, fmt.Sprintf("fail to exec command info:sentinel"))
	}

	// extract addr/port from names and addrs
	for i := 0; i < len(ips); i++ {
		var (
			tmpStrs []string
		)
		equalReg := regexp.MustCompile(`=`)

		// split ip string and get ip
		tmpStrs = equalReg.Split(ips[i], -1) // use '=' to split the ip string
		slave.IP = tmpStrs[1]

		// split port string and get port
		tmpStrs = equalReg.Split(ports[i], -1)
		slave.Port, _ = strconv.Atoi(tmpStrs[1])

		// insert slave to slaves
		slaves = append(slaves, slave)
	}

	return slaves, nil
}

// get sentinel info once
func getSentinelInfo() ([]Meta, error) {
	var (
		err        error
		metas      []Meta
		nameReg    *regexp.Regexp
		addressReg *regexp.Regexp
		names      []string
		addrs      []string
	)

	// connect redis
	conn, err := redis.Dial("tcp", SentinelAddr)
	if err != nil {
		return metas, errors.Wrapf(err, fmt.Sprintf("fail to connect sentinel %s", SentinelAddr))
	}
	defer conn.Close()

	// get result string
	allStr, err := redis.String(conn.Do("info", "Sentinel"))
	if err != nil {
		return metas, errors.Wrapf(err, fmt.Sprintf("fail to exec command info:sentinel"))
	}

	// parse name, using regular expression
	nameReg = regexp.MustCompile(`(master[0-9]{1,}):name=([^0-9]{1,})[0-9]`)
	names = nameReg.FindAllString(allStr, -1)

	// parse address:port, using regular expression
	addressReg = regexp.MustCompile(`address=(([0-9]{1,}.){3})([0-9]{1,}):([0-9]{1,})`)
	addrs = addressReg.FindAllString(allStr, -1)

	// if name count != addr count, return error
	if len(names) != len(addrs) {
		return metas, errors.Wrapf(err, fmt.Sprintf("fail to exec command info:sentinel"))
	}

	// extract name/addr/port from names and addrs
	for i := 0; i < len(names); i++ {
		var (
			tmpMeta Meta
			tmpStrs []string
		)
		equalReg := regexp.MustCompile(`=`)
		colonReg := regexp.MustCompile(`:`)

		// split name string and get master name
		tmpStrs = equalReg.Split(names[i], -1) // use ‘=’ to split the name string
		tmpMeta.Name = tmpStrs[1]

		// split address string and get master addr(ip, port)
		tmpStrs = equalReg.Split(addrs[i], -1)   // use '=' to split the address string
		tmpStrs = colonReg.Split(tmpStrs[1], -1) // use ':' to split the sub string
		tmpMeta.Master.IP = tmpStrs[0]
		tmpMeta.Master.Port, _ = strconv.Atoi(tmpStrs[1])

		// get slave info
		tmpMeta.Slaves, err = getSlavesInfo(tmpMeta.Master)
		if err != nil {
			Log.Warn("failed to get redis slave info, error:%#v", err)
		}

		// insert theMeta to metas
		metas = append(metas, tmpMeta)
	}

	return metas, nil
}

// get sentinel info periodically
func GetSentinelInfo(wh *gxtime.Wheel, span time.Duration) {
	var (
		err   error
		metas []Meta
		str   string
		ver   int32
	)

	for {
		select {
		case <-wh.After(span):
			metas, err = getSentinelInfo()
			if err != nil {
				Log.Warn("failed to get redis cluster info, error:%#v", err)
				break
			}
			ver = atomic.AddInt32(&version, 1)
			infoMap[int(ver)] = metas
			SetMeta(int(ver), metas)
			delete(infoMap, int(ver-1024))
			DelMeta(int(ver-1024))

		case str = <-ch:
			// get latest metas info because of received +switch-master message
			ver = atomic.LoadInt32(&version)
			metas = infoMap[int(ver)]

			// use ‘ ’ to split the name string
			// split: 0-master name;
			//        1-old master ip; 2-old master port
			//        3-new master ip; 4-new master port
			spaceReg := regexp.MustCompile(` `)
			splitStrs := spaceReg.Split(str, -1)

			// if the count of split string is less than 5, skip
			if len(splitStrs) < 5 {
				continue
			}

			// if convert from string to int returns error, skip
			newMasterPort, err := strconv.Atoi(splitStrs[4])
			if err != nil {
				continue
			}

			// find the right master-slave instance`s meta info
			for i, _ := range metas {
				if strings.Compare(metas[i].Name, splitStrs[0]) == 0 {
					// update master info with the new promoted node info
					metas[i].Master.IP = splitStrs[3]
					metas[i].Master.Port = newMasterPort

					// delete the slave info which promoted to master with old master info
					for j, _ := range metas[i].Slaves {
						if strings.Compare(metas[i].Slaves[j].IP, splitStrs[3]) == 0 && metas[i].Slaves[j].Port == newMasterPort {
							metas[i].Slaves = append(metas[i].Slaves[:j], metas[i].Slaves[j+1:]...)
							break
						}
					}
					break
				} // end if
			} // end for

			ver = atomic.AddInt32(&version, 1)
			infoMap[int(ver)] = metas
			SetMeta(int(ver), metas)
			delete(infoMap, int(ver-1024))
			DelMeta(int(ver-1024))
		}
	}
}

func SubscribeODown() {

	// connect redis
	conn, err := redis.Dial("tcp", SentinelAddr)
	if err != nil {
		fmt.Println("connect error: ", err)
		return
	}
	defer conn.Close()

	psc := redis.PubSubConn{conn}
	psc.Subscribe("+switch-master")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			ch <- fmt.Sprintf("%s", v.Data)
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
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

	infoMap = make(map[int][]Meta, 1024)

	fmt.Println("start listen...")

	// create a new wheel
	wh = gxtime.NewWheel(gxtime.TimeMillisecondDuration(1000), 20)

	go SubscribeODown()

	// get sentinel info periodically with timewheel
	GetSentinelInfo(wh, gxtime.TimeMillisecondDuration(3000))
}

func SetMeta(ver int, data []Meta) error {
	// connect redis
	conn, err := redis.Dial("tcp", DBAddr)
	if err != nil {
		fmt.Println("connect error: ", err)
		return err
	}
	defer conn.Close()

	// convert data from Meta to byte
	value, err := json.Marshal(data)
	if err != nil {
		Log.Warn("json marshal err,%s", err)
		return err
	}

	_, err = conn.Do("SET", ver, value)
	if err != nil {
		fmt.Println("store error\n")
		return err
	}

	return nil
}

func DelMeta(ver int) (err error) {

	// connect redis
	conn, err := redis.Dial("tcp", DBAddr)
	if err != nil {
		fmt.Println("connect error: ", err)
		return err
	}
	defer conn.Close()

	// delete data which version is ver
	_, err = conn.Do("DEL", ver)
	if err != nil {
		fmt.Println("delete error\n")
		return err
	}

	return nil
}

