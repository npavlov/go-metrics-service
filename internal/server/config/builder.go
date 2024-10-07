package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/flags"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

// Builder defines the builder for the Config struct
type Builder struct {
	cfg *Config
}

// NewConfigBuilder initializes the ConfigBuilder with default values
func NewConfigBuilder() *Builder {
	return &Builder{
		cfg: &Config{},
	}
}

// FromEnv parses environment variables into the ConfigBuilder
func (b *Builder) FromEnv() *Builder {
	if err := env.Parse(b.cfg); err != nil {
		fmt.Printf("Error parsing environment variables: %+v\n", err)
	}
	return b
}

// FromFlags parses command line flags into the ConfigBuilder
func (b *Builder) FromFlags() *Builder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to run server")
	flag.Parse()
	flags.VerifyFlags()
	return b
}

// Build returns the final configuration
func (b *Builder) Build() *Config {
	return b.cfg
}
