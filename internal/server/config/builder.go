package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/logger"
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
	l := logger.Get()

	if err := env.Parse(b.cfg); err != nil {
		l.Error().Err(err).Msg("failed to parse environment variables")
	}
	return b
}

// FromFlags parses command line flags into the ConfigBuilder
func (b *Builder) FromFlags() *Builder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to run server")
	flag.Parse()
	return b
}

// Build returns the final configuration
func (b *Builder) Build() *Config {

	return b.cfg
}
