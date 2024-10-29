package helpers

import "strings"

func IsCompressible(contentType string) bool {
	compressibleTypes := map[string]bool{
		"application/javascript": true,
		"application/json":       true,
		"text/css":               true,
		"text/html":              true,
		"text/plain":             true,
		"text/xml":               true,
		"application/text":       true,
	}

	for ct := range compressibleTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}

	return false
}
