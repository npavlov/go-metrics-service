package middlewares

import (
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

type compressBrWriter struct {
	w  http.ResponseWriter
	zw *brotli.Writer
}

func newCompressBrWriter(w http.ResponseWriter) *compressBrWriter {
	return &compressBrWriter{
		w:  w,
		zw: brotli.NewWriter(w),
	}
}

// Header - the function that retrieves header from the response writer.
func (c *compressBrWriter) Header() http.Header {
	return c.w.Header()
}

// Write - the function that writes to the response.
func (c *compressBrWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to close writer")
	}

	return n, nil
}

// WriteHeader  - the function that writes header to the response.
func (c *compressBrWriter) WriteHeader(statusCode int) {
	if statusCode <= http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close - the function that closes the response writer.
func (c *compressBrWriter) Close() error {
	if err := c.zw.Close(); err != nil {
		return errors.Wrapf(err, "failed to close writer")
	}

	return nil
}

// BrotliMiddleware - the middleware function that enables compression for http communicztion.
func BrotliMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsBrotli := strings.Contains(acceptEncoding, "br")
		if supportsBrotli {
			cw := newCompressBrWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					return
				}
			}()
			w.Header().Set("Content-Encoding", "br")
		}

		next.ServeHTTP(ow, r)
	})
}
