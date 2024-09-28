package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

// Config Only starting with upper case
type Config struct {
	Address string `env:"ADDRESS"`
}

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	if len(cfg.Address) > 0 {
		flagRunAddr = cfg.Address
	}
}
