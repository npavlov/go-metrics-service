//nolint:tagliatelle
package config

import (
	"flag"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/utils"
)

type Config struct {
	Address           string `env:"ADDRESS"         envDefault:"localhost:8080" json:"address"`
	Key               string `env:"KEY"             envDefault:""               json:"key"`
	CryptoKey         string `env:"CRYPTO_KEY"      envDefault:""               json:"crypto_key"`
	ReportInterval    int64  `env:"REPORT_INTERVAL" envDefault:"10"             json:"report_interval"`
	PollInterval      int64  `env:"POLL_INTERVAL"   envDefault:"2"              json:"poll_interval"`
	UseBatch          bool   `env:"USE_BATCH"       envDefault:"false"          json:"use_batch"`
	RateLimit         int    `env:"RATE_LIMIT"      envDefault:"10"             json:"rate_limit"`
	Config            string `env:"CONFIG"          envDefault:""`
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
			Key:               "",
			RateLimit:         0,
			CryptoKey:         "",
			Config:            "",
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
	flag.StringVar(&b.cfg.Key, "k", b.cfg.Key, "key to sign request")
	flag.StringVar(&b.cfg.CryptoKey, "crypto-key", b.cfg.CryptoKey, "crypto key to sign request")
	flag.IntVar(&b.cfg.RateLimit, "l", b.cfg.RateLimit, "rate limit for workers")
	flag.StringVar(&b.cfg.Config, "config", b.cfg.Config, "path to config file")
	flag.Parse()

	return b
}

// FromObj sets cfg from object.
func (b *Builder) FromObj(cfg *Config) *Builder {
	b.cfg = cfg

	return b
}

// FromFile sets cfg from config file.
func (b *Builder) FromFile() *Builder {
	//nolint:exhaustruct
	newConfig := &Config{}
	err := utils.ReadFromFile(b.cfg.Config, newConfig, b.logger)
	if err != nil {
		b.logger.Error().Err(err).Msg("failed to read config file")
	}

	utils.ReplaceValues(newConfig, b.cfg)

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
