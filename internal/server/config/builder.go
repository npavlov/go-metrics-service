package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
)

type Config struct {
	Address          string `env:"ADDRESS"        envDefault:"localhost:8080"`
	StoreInterval    int64  `env:"STORE_INTERVAL" envDefault:"300"`
	StoreIntervalDur time.Duration
	File             string `env:"FILE_STORAGE_PATH"     envDefault:"temp.txt"`
	Database         string `env:"DATABASE_DSN"          envDefault:""`
	RestoreStorage   bool   `env:"RESTORE"               envDefault:"true"`
	HealthCheck      int64  `env:"HEALTH_CHECK_INTERVAL" envDefault:"5"`
	Key              string `env:"KEY"                   envDefault:""`
	HealthCheckDur   time.Duration
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
			Address:          "",
			StoreInterval:    0,
			File:             "",
			RestoreStorage:   false,
			StoreIntervalDur: 0,
			Database:         "",
			HealthCheck:      0,
			HealthCheckDur:   0,
			Key:              "",
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
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to run server")
	flag.BoolVar(&b.cfg.RestoreStorage, "r", b.cfg.RestoreStorage, "restore previous session")
	flag.StringVar(&b.cfg.File, "f", b.cfg.File, "file where to store mem storage")
	flag.StringVar(&b.cfg.Database, "d", b.cfg.Database, "database DSN")
	flag.Int64Var(&b.cfg.StoreInterval, "i", b.cfg.StoreInterval, "time flushing mem storage to file (in seconds)")
	flag.StringVar(&b.cfg.Key, "k", b.cfg.Key, "key to sign request")
	flag.Parse()

	return b
}

// Build returns the final configuration.
func (b *Builder) Build() *Config {
	b.cfg.StoreIntervalDur = time.Duration(b.cfg.StoreInterval) * time.Second
	b.cfg.HealthCheckDur = time.Duration(b.cfg.HealthCheck) * time.Second

	return b.cfg
}
