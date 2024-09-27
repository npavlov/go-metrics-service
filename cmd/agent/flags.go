package main

import (
	"flag"
	"time"
)

var flagRunAddr string
var pollInterval time.Duration
var reportInterval time.Duration

func parseFlags() {

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.DurationVar(&reportInterval, "r", 10, "report interval to send metrics")
	flag.DurationVar(&pollInterval, "p", 2, "poll interval to update metrics")
	flag.Parse()
}
