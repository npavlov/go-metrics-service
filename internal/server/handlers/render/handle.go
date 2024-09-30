package render

import (
	"embed"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var tplFolder embed.FS

type MetricsPage struct {
	Gauges   map[types.MetricName]float64
	Counters map[types.MetricName]int64
}

func GetRenderHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		page := MetricsPage{
			Gauges:   ms.GetGauges(),
			Counters: ms.GetCounters(),
		}

		// Parse the template files from the embedded filesystem
		tmpl, err := template.ParseFS(tplFolder, "templates/*.html")
		if err != nil {
			log.Printf("Error parsing templates: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, page); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
