package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/flags"
)

// ConfigBuilder defines the builder for the Config struct
type ConfigBuilder struct {
	cfg *config.Config
}

// NewConfigBuilder initializes the ConfigBuilder with default values
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		cfg: &config.Config{},
	}
}

// FromEnv parses environment variables into the ConfigBuilder
func (b *ConfigBuilder) FromEnv() *ConfigBuilder {
	if err := env.Parse(b.cfg); err != nil {
		fmt.Printf("Error parsing environment variables: %+v\n", err)
	}
	return b
}

// FromFlags parses command line flags into the ConfigBuilder
func (b *ConfigBuilder) FromFlags() *ConfigBuilder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to reach server")
	flag.Int64Var(&b.cfg.ReportInterval, "r", b.cfg.ReportInterval, "report interval to send watcher")
	flag.Int64Var(&b.cfg.PollInterval, "p", b.cfg.PollInterval, "poll interval to update watcher")
	flag.Parse()
	flags.VerifyFlags()
	return b
}

// Build returns the final configuration
func (b *ConfigBuilder) Build() *config.Config {
	return b.cfg
}
