package handlers

import (
	_ "embed"
	"html/template"
	"net/http"
	"sort"
)

//go:embed templates/index.html
var indexHTML string

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
		tmpl, err := template.New("index").Parse(indexHTML)
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
