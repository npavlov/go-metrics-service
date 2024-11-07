package config

import (
	"flag"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
)

type Config struct {
	Address           string `env:"ADDRESS"         envDefault:"localhost:8080"`
	ReportInterval    int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval      int64  `env:"POLL_INTERVAL"   envDefault:"2"`
	UseBatch          bool   `env:"USE_BATCH"       envDefault:"false"`
	ReportIntervalDur time.Duration
	PollIntervalDur   time.Duration
}

// Builder defines the builder for the Config struct.
type Builder struct {
	cfg    *Config
	logger *zerolog.Logger
}

// NewConfigBuilder initializes the ConfigBuilder with default values.
func NewConfigBuilder(log *zerolog.Logger) *Builder {
	return &Builder{
		cfg: &Config{
			Address:           "",
			ReportInterval:    0,
			PollInterval:      0,
			PollIntervalDur:   0,
			ReportIntervalDur: 0,
			UseBatch:          false,
		},
		logger: log,
	}
}

// FromEnv parses environment variables into the ConfigBuilder.
func (b *Builder) FromEnv() *Builder {
	if err := env.Parse(b.cfg); err != nil {
		b.logger.Error().Err(err).Msg("failed to parse environment variables")
	}

	return b
}

// FromFlags parses command line flags into the ConfigBuilder.
func (b *Builder) FromFlags() *Builder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to reach server")
	flag.Int64Var(&b.cfg.ReportInterval, "r", b.cfg.ReportInterval, "report interval to send watcher (in seconds)")
	flag.Int64Var(&b.cfg.PollInterval, "p", b.cfg.PollInterval, "poll interval to update watcher (in seconds)")
	flag.Parse()

	return b
}

// FromObj sets cfg from object.
func (b *Builder) FromObj(cfg *Config) *Builder {
	b.cfg = cfg

	return b
}

// Build returns the final configuration.
func (b *Builder) Build() *Config {
	if !strings.HasPrefix(b.cfg.Address, "http://") && !strings.HasPrefix(b.cfg.Address, "https://") {
		b.cfg.Address = "http://" + b.cfg.Address
	}
	b.cfg.PollIntervalDur = time.Duration(b.cfg.PollInterval) * time.Second
	b.cfg.ReportIntervalDur = time.Duration(b.cfg.ReportInterval) * time.Second

	return b.cfg
}
