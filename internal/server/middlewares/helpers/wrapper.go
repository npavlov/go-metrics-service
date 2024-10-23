package helpers

import (
	"io"
	"net/http"
)

type WrappedResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (wr *WrappedResponseWriter) Write(bytes []byte) (int, error) {
	// Only compress text, html, or json content types
	contentType := wr.Header().Get("Content-Type")
	if IsCompressible(contentType) {
		return wr.Writer.Write(bytes)
	}

	// Write normally if content is not compressible
	return wr.ResponseWriter.Write(bytes)
}
