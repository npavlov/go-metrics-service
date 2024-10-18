package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header - the function that retrieves header from the response writer.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write - the function that writes to the response.
func (c *compressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to close writer")
	}

	return n, nil
}

// WriteHeader  - the function that writes header to the response.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode <= http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close - the function that closes the response writer.
func (c *compressWriter) Close() error {
	if err := c.zw.Close(); err != nil {
		return errors.Wrapf(err, "failed to close writer")
	}

	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader function to create a new gzip reader.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	// Create a new gzip reader from the input ReadCloser
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read method to satisfy the io.Reader interface.
func (cr *compressReader) Read(p []byte) (int, error) {
	return cr.zr.Read(p)
}

// Close method to satisfy the io.Closer interface.
func (cr *compressReader) Close() error {
	// Close both the gzip reader and the original ReadCloser
	if err := cr.zr.Close(); err != nil {
		return err
	}

	return cr.r.Close()
}

// GzipMiddleware - the middleware function that enables compression for http communicztion.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		supportsBrotli := strings.Contains(acceptEncoding, "br")
		if supportsGzip && !supportsBrotli {
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					return
				}
			}()
			w.Header().Set("Content-Encoding", "gzip")
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer func(cr *compressReader) {
				_ = cr.Close()
			}(cr)
		}

		next.ServeHTTP(ow, r)
	})
}
