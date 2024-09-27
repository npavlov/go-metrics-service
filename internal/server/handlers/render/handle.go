package render

import (
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"html/template"
	"net/http"
)

// Define the template for rendering the HTML page.
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
    <style>
        table { border-collapse: collapse; width: 50%; margin: 20px auto; }
        th, td { border: 1px solid black; padding: 8px; text-align: center; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1 style="text-align: center;">Metrics</h1>
    <table>
        <thead>
            <tr>
                <th>Metric Name</th>
                <th>Values</th>
            </tr>
        </thead>
        <tbody>
            {{range $name, $value := .Counters}}
                <tr>
                    <td>{{$name}}</td>
                    <td>{{$value}}</td>
                </tr>
            {{end}}
            {{range $name, $value := .Gauges}}
                <tr>
                    <td>{{$name}}</td>
                    <td>{{$value}}</td>
                </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`

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

		tmpl := template.Must(template.New("metrics").Parse(htmlTemplate))
		if err := tmpl.Execute(w, page); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
