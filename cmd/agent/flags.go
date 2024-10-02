package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
)

func parseFlags() *config.Config {
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Error parsing environment variables: %+v\n", err)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")
	flag.Int64Var(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval to send watcher")
	flag.Int64Var(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval to update watcher")
	flag.Parse()

	return &config.Config{Address: cfg.Address, ReportInterval: cfg.ReportInterval, PollInterval: cfg.PollInterval}
}
