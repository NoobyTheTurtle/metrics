package handlers

import (
	"html/template"
	"net/http"
	"sort"
)

const indexTemplate = `
<!DOCTYPE html>
<html>
<head></head>
<body>
    <h1>Metrics</h1>
    
    <h2>Gauge</h2>
    <table>
        <tr>
            <th>Name</th>
            <th>Value</th>
        </tr>
        {{range .Gauges}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{.Value}}</td>
        </tr>
        {{end}}
    </table>
    
    <h2>Counter</h2>
    <table>
        <tr>
            <th>Name</th>
            <th>Value</th>
        </tr>
        {{range .Counters}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{.Value}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`

type metricData struct {
	Name  string
	Value any
}

type pageData struct {
	Gauges   []metricData
	Counters []metricData
}

func mapGauges(gauges map[string]float64) []metricData {
	result := make([]metricData, 0, len(gauges))
	for name, value := range gauges {
		result = append(result, metricData{
			Name:  name,
			Value: value,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func mapCounters(counters map[string]int64) []metricData {
	result := make([]metricData, 0, len(counters))
	for name, value := range counters {
		result = append(result, metricData{
			Name:  name,
			Value: value,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (h *handler) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(indexTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := pageData{
			Gauges:   mapGauges(h.storage.GetAllGauges()),
			Counters: mapCounters(h.storage.GetAllCounters()),
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
