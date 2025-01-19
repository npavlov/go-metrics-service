package web

import (
	"embed"
	"html/template"
	"sync"

	"github.com/pkg/errors"
)

//go:embed template/*
var tplFolder embed.FS

type Reader interface {
	Read(filename string) (*template.Template, error)
}

type EmbedReader struct {
	cache map[string]*template.Template
	mu    sync.RWMutex // To ensure thread-safe access to the cache
}

// NewEmbedReader - constructor for EmbedReader.
func NewEmbedReader() *EmbedReader {
	return &EmbedReader{
		cache: make(map[string]*template.Template),
	}
}

func (t *EmbedReader) Read(filename string) (*template.Template, error) {
	// Check the cache first
	t.mu.RLock()
	if tmpl, found := t.cache[filename]; found {
		t.mu.RUnlock()
		return tmpl, nil
	}
	t.mu.RUnlock()

	// Load the template if not found in the cache
	tmpl, err := template.ParseFS(tplFolder, "template/"+filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse template")
	}

	// Store the template in the cache
	t.mu.Lock()
	t.cache[filename] = tmpl
	t.mu.Unlock()

	return tmpl, nil
}
