package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

import (
	"github.com/AlexStocks/goext/log"
	"github.com/AlexStocks/goext/net"
	"github.com/AlexStocks/goext/time"
)

const (
	APP_CONF_FILE     string = "APP_CONF_FILE"
	APP_LOG_CONF_FILE string = "APP_LOG_CONF_FILE"
)

const (
	FailfastTimeout = 3 // in second
)

var (
	pprofPath = "/debug/pprof/"

	usageStr = `
Usage: log-kafka [options]
Server Options:
    -c, --config <file>              Configuration file path
    -l, --log <file>                 Log configuration file
Common Options:
    -h, --help                       Show this message
    -v, --version                    Show version
`
)

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func getHostInfo() {
	var (
		err error
	)

	LocalHost, err = os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("os.Hostname() = %s", err))
	}

	LocalIP, err = gxnet.GetLocalIP(LocalIP)
	if err != nil {
		panic("can not get local IP!")
	}

	ProcessID = fmt.Sprintf("%s@%s", LocalIP, LocalHost)
}

func createPIDFile() error {
	if !Conf.Core.PID.Enabled {
		return nil
	}

	pidPath := Conf.Core.PID.Path
	_, err := os.Stat(pidPath)
	if os.IsNotExist(err) || Conf.Core.PID.Override {
		currentPid := os.Getpid()
		if err := os.MkdirAll(filepath.Dir(pidPath), os.ModePerm); err != nil {
			return fmt.Errorf("Can't create PID folder on %v", err)
		}

		file, err := os.Create(pidPath)
		if err != nil {
			return fmt.Errorf("Can't create PID file: %v", err)
		}
		defer file.Close()
		if _, err := file.WriteString(fmt.Sprintf("%s-%s", ProcessID, strconv.FormatInt(int64(currentPid), 10))); err != nil {
			return fmt.Errorf("Can'write PID information on %s: %v", pidPath, err)
		}
	} else {
		return fmt.Errorf("%s already exists", pidPath)
	}
	return nil
}

// initLog use for initial log module
func initLog(logConf string) {
	Log = gxlog.NewLoggerWithConfFile(logConf)
	Log.SetAsDefaultLogger()
}

func initSignal() {
	var (
		// signal.Notify的ch信道是阻塞的(signal.Notify不会阻塞发送信号), 需要设置缓冲
		signals = make(chan os.Signal, 1)
		ticker  = time.NewTicker(gxtime.TimeSecondDuration(Conf.Redis.UpdateInterval))
	)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case sig := <-signals:
			Log.Info("get signal %s", sig.String())
			switch sig {
			case syscall.SIGHUP:
			// reload()
			default:
				go gxtime.Future(Conf.Core.FailFastTimeout, func() {
					Log.Warn("app exit now by force...")
					Log.Close()
					os.Exit(1)
				})

				// 要么survialTimeout时间内执行完毕下面的逻辑然后程序退出，要么执行上面的超时函数程序强行退出
				worker.Close()
				Log.Warn("app exit now...")
				Log.Close()
				return
			}

		case <-ticker.C:
			worker.updateClusterMeta()
		}
	}
}

func main() {
	var (
		err         error
		showVersion bool
		configFile  string
		logConf     string
	)

	/////////////////////////////////////////////////
	// conf
	/////////////////////////////////////////////////

	SetVersion(Version)

	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.StringVar(&configFile, "c", "", "Configuration file path.")
	flag.StringVar(&configFile, "config", "", "Configuration file path.")
	flag.StringVar(&logConf, "l", "", "Logger configuration file.")
	flag.StringVar(&logConf, "log", "", "Logger configuration file.")

	flag.Usage = usage
	flag.Parse()

	// Show version and exit
	if showVersion {
		PrintVersion()
		os.Exit(0)
	}

	if configFile == "" {
		configFile = os.Getenv(APP_CONF_FILE)
		if configFile == "" {
			panic("can not get configFile!")
		}
	}
	if path.Ext(configFile) != ".yml" {
		panic(fmt.Sprintf("application configure file name{%v} suffix must be .yml", configFile))
	}
	Conf, err = LoadConfYaml(configFile)
	if err != nil {
		log.Printf("Load yaml config file error: '%v'", err)
		return
	}
	fmt.Printf("config: %+v\n", Conf)

	if logConf == "" {
		logConf = os.Getenv(APP_LOG_CONF_FILE)
		if logConf == "" {
			panic("can not get logConf!")
		}
	}

	/////////////////////////////////////////////////
	// worker
	/////////////////////////////////////////////////
	if Conf.Core.FailFastTimeout == 0 {
		Conf.Core.FailFastTimeout = FailfastTimeout
	}

	getHostInfo()

	initLog(logConf)

	if err = createPIDFile(); err != nil {
		Log.Critic(err)
	}
	worker = NewSentinelWorker()
	if err = worker.WatchInstanceSwitch(); err != nil {
		panic(fmt.Sprintf("failed to start watch instance switch goroutine, error:%#v", err))
	}

	initSignal()
}
