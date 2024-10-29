package testutils

import (
	"os"

	"github.com/rs/zerolog"
)

func GetTLogger() *zerolog.Logger {
	l := zerolog.New(os.Stdout)

	return &l
}
