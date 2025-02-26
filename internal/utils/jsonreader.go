package utils

import (
	"io"
	"os"
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// ReadFromFile reads JSON into any provided structure.
func ReadFromFile[T any](file string, temp *T, logger *zerolog.Logger) error {
	// Open JSON file
	jsonFile, err := os.Open(file)
	if err != nil {
		logger.Error().Err(err).Msg("failed to open config file")

		return errors.Wrap(err, "failed to open config file")
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	// Read file content
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read config file")

		return errors.Wrap(err, "failed to read config file")
	}

	// JSON parsing
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err = json.Unmarshal(byteValue, temp)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse config file")

		return errors.Wrap(err, "failed to parse config file")
	}

	return nil
}

// ReplaceValues copies non null values to target.
func ReplaceValues[T any](source T, target T) {
	srcVal := reflect.ValueOf(source).Elem()
	tgtVal := reflect.ValueOf(target).Elem()

	for i := range srcVal.NumField() {
		srcField := srcVal.Field(i)
		tgtField := tgtVal.Field(i)

		// Verify that field can be set
		if !tgtField.CanSet() {
			continue
		}

		// Copy value only if source != zero (non-null)
		if !isZeroValue(srcField) {
			tgtField.Set(srcField)
		}
	}
}

// isZeroValue verifies whether the field is empty or not.
func isZeroValue(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
