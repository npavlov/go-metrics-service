package main

import (
	"io"
	"log/slog"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"github.com/npavlov/go-metrics-service/internal/model"
)

func main() {
	//nolint:exhaustruct
	stmts, err := gormschema.New("postgres").Load(&model.Metric{})
	if err != nil {
		slog.Error("Failed to load gorm schema", "err", err.Error())
		os.Exit(1)
	}
	_, err = io.WriteString(os.Stdout, stmts)
	if err != nil {
		slog.Error("Failed to write gorm schema", "err", err.Error())
		os.Exit(1)
	}
}
