package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

// Config Only starting with upper case
type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func parseFlags() *Config {
	var flagRunAddr string

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Error parsing environment variables: %+v\n", err)
	}

	flag.StringVar(&flagRunAddr, "a", cfg.Address, "address and port to run server")
	flag.Parse()

	return &Config{flagRunAddr}
}
