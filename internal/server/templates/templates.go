package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed files/*
var tplFolder embed.FS

type Reader interface {
	Read(filename string) (*template.Template, error)
}

type EmbedReader struct {
}

// NewEmbedReader - constructor for MetricsHandler
func NewEmbedReader() *EmbedReader {
	return &EmbedReader{}
}

func (t *EmbedReader) Read(filename string) (*template.Template, error) {
	tmpl, err := template.ParseFS(tplFolder, fmt.Sprintf("files/%s", filename))
	if err != nil {
		return nil, err
	}

	return tmpl, err
}
