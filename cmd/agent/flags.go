package main

import (
	"flag"
)

var flagRunAddr string
var pollInterval int64
var reportInterval int64

func parseFlags() {

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "report interval to send metrics")
	flag.Int64Var(&pollInterval, "p", 2, "poll interval to update metrics")
	flag.Parse()
}
