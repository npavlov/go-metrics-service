//nolint:tagliatelle
package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/utils"
)

type Config struct {
	Address          string `env:"ADDRESS"        envDefault:"localhost:8080" json:"address"`
	GRPCAddress      string `env:"GRPC_ADDRESS"   envDefault:":9090"          json:"grpc_address"`
	StoreInterval    int64  `env:"STORE_INTERVAL" envDefault:"300"            json:"store_interval"`
	StoreIntervalDur time.Duration
	File             string `env:"FILE_STORAGE_PATH"     envDefault:"temp.txt" json:"store_file"`
	Database         string `env:"DATABASE_DSN"          envDefault:""         json:"database_dsn"`
	RestoreStorage   bool   `env:"RESTORE"               envDefault:"true"     json:"restore"`
	HealthCheck      int64  `env:"HEALTH_CHECK_INTERVAL" envDefault:"5"        json:"health_check_interval"`
	Key              string `env:"KEY"                   envDefault:""         json:"key"`
	CryptoKey        string `env:"CRYPTO_KEY"            envDefault:""         json:"crypto_key"`
	TrustedSubnet    string `env:"TRUSTED_SUBNET"        envDefault:""         json:"trusted_subnet"`
	Config           string `env:"CONFIG"                envDefault:""`
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
			GRPCAddress:      "",
			StoreInterval:    0,
			File:             "",
			RestoreStorage:   false,
			StoreIntervalDur: 0,
			Database:         "",
			HealthCheck:      0,
			HealthCheckDur:   0,
			Key:              "",
			CryptoKey:        "",
			Config:           "",
			TrustedSubnet:    "",
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
	flag.StringVar(&b.cfg.GRPCAddress, "grpca", b.cfg.GRPCAddress, "address and port to run server using gRPC")
	flag.BoolVar(&b.cfg.RestoreStorage, "r", b.cfg.RestoreStorage, "restore previous session")
	flag.StringVar(&b.cfg.File, "f", b.cfg.File, "file where to store mem storage")
	flag.StringVar(&b.cfg.Database, "d", b.cfg.Database, "database DSN")
	flag.Int64Var(&b.cfg.StoreInterval, "i", b.cfg.StoreInterval, "time flushing mem storage to file (in seconds)")
	flag.StringVar(&b.cfg.Key, "k", b.cfg.Key, "key to sign request")
	flag.StringVar(&b.cfg.CryptoKey, "crypto-key", b.cfg.CryptoKey, "crypto key to sign request")
	flag.StringVar(&b.cfg.TrustedSubnet, "t", b.cfg.TrustedSubnet, "trusted subnet")
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
	b.cfg.StoreIntervalDur = time.Duration(b.cfg.StoreInterval) * time.Second
	b.cfg.HealthCheckDur = time.Duration(b.cfg.HealthCheck) * time.Second

	return b.cfg
}
