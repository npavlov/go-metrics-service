package templates

import (
	"embed"
	"html/template"

	"github.com/pkg/errors"
)

//go:embed files/*
var tplFolder embed.FS

type Reader interface {
	Read(filename string) (*template.Template, error)
}

type EmbedReader struct{}

// NewEmbedReader - constructor for MetricsHandler.
func NewEmbedReader() *EmbedReader {
	return &EmbedReader{}
}

func (t *EmbedReader) Read(filename string) (*template.Template, error) {
	tmpl, err := template.ParseFS(tplFolder, "files/"+filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse template")
	}

	return tmpl, nil
}
