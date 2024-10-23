package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
)

type Config struct {
	Address        string `env:"ADDRESS"           envDefault:"localhost:8080"`
	StoreInterval  int64  `env:"STORE_INTERVAL"    envDefault:"300"`
	File           string `env:"FILE_STORAGE_PATH" envDefault:"temp.txt"`
	RestoreStorage bool   `env:"RESTORE"           envDefault:"true"`
}

// Builder defines the builder for the Config struct.
type Builder struct {
	cfg *Config
	l   *zerolog.Logger
}

// NewConfigBuilder initializes the ConfigBuilder with default values.
func NewConfigBuilder(log *zerolog.Logger) *Builder {
	return &Builder{
		cfg: &Config{
			Address:        "",
			StoreInterval:  0,
			File:           "",
			RestoreStorage: false,
		},
		l: log,
	}
}

// FromEnv parses environment variables into the ConfigBuilder.
func (b *Builder) FromEnv() *Builder {
	if err := env.Parse(b.cfg); err != nil {
		b.l.Error().Err(err).Msg("failed to parse environment variables")
	}

	return b
}

// FromFlags parses command line flags into the ConfigBuilder.
func (b *Builder) FromFlags() *Builder {
	flag.StringVar(&b.cfg.Address, "a", b.cfg.Address, "address and port to run server")
	flag.BoolVar(&b.cfg.RestoreStorage, "r", b.cfg.RestoreStorage, "restore previous session")
	flag.StringVar(&b.cfg.File, "f", b.cfg.File, "file where to store mem storage")
	flag.Parse()

	return b
}

// Build returns the final configuration.
func (b *Builder) Build() *Config {
	return b.cfg
}
