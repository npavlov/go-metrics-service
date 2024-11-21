package utils

import (
	"compress/gzip"
	"io"

	"github.com/pkg/errors"
)

func DecompressResult(body io.ReadCloser) ([]byte, error) {
	// Decompress the gzipped response
	reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress result")
	}
	defer func(reader *gzip.Reader) {
		_ = reader.Close()
	}(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress result")
	}

	return data, nil
}
