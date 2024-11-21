package utils

import (
	"bytes"
	"compress/gzip"

	"github.com/pkg/errors"
)

func Compress(data []byte) (*bytes.Buffer, error) {
	var compressedPayload bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedPayload)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress data")
	}
	if err = gzipWriter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close gzip writer")
	}

	return &compressedPayload, nil
}
