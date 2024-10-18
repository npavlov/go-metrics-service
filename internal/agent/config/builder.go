package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/logger"
)

type Config struct {
	Address        string `env:"ADDRESS"         envDefault:"localhost:8080"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int64  `env:"POLL_INTERVAL"   envDefault:"2"`
}

// Builder defines the builder for the Config struct.
type Builder struct {
	cfg *Config
}

// NewConfigBuilder initializes the ConfigBuilder with default values.
func NewConfigBuilder() *Builder {
	return &Builder{
		cfg: &Config{},
	}
}

// FromEnv parses environment variables into the ConfigBuilder.
func (b *Builder) FromEnv() *Builder {
	l := logger.NewLogger().Get()

	if err := env.Parse(b.cfg); err != nil {
		l.Error().Err(err).Msg("failed to parse environment variables")
	}

	return b
}

// FromFlags parses command line flags into the ConfigBuilder.
func (b *Builder) FromFlags() *Builder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to reach server")
	flag.Int64Var(&b.cfg.ReportInterval, "r", b.cfg.ReportInterval, "report interval to send watcher")
	flag.Int64Var(&b.cfg.PollInterval, "p", b.cfg.PollInterval, "poll interval to update watcher")
	flag.Parse()

	return b
}

// Build returns the final configuration.
func (b *Builder) Build() *Config {
	if !strings.HasPrefix(b.cfg.Address, "http://") && !strings.HasPrefix(b.cfg.Address, "https://") {
		b.cfg.Address = "http://" + b.cfg.Address
	}

	return b.cfg
}
