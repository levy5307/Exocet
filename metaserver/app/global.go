package main

import (
	"github.com/AlexStocks/goext/log"
)

type (
	empty interface{}
)

var (
	// local ip
	LocalIP   string
	LocalHost string
	// progress id
	ProcessID string
	// Conf is main config
	Conf ConfYaml
	// Log records server request log
	Log gxlog.Logger
	// worker
	worker *SentinelWorker
)
