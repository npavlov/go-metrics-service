package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

var flagRunAddr string
var pollInterval int64
var reportInterval int64

// Config Only starting with upper case
type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "report interval to send metrics")
	flag.Int64Var(&pollInterval, "p", 2, "poll interval to update metrics")
	flag.Parse()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	if len(cfg.Address) > 0 {
		flagRunAddr = cfg.Address
	}
	if cfg.ReportInterval > 0 {
		reportInterval = cfg.ReportInterval
	}
	if cfg.PollInterval > 0 {
		pollInterval = cfg.PollInterval
	}
}
