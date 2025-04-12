package html

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

type IndexStorage interface {
	GaugesGetter
	CountersGetter
}

type indexHandler struct {
	storage IndexStorage
}

func newIndexHandler(storage IndexStorage) *indexHandler {
	return &indexHandler{
		storage: storage,
	}
}

func mapMetrics[T any](metrics map[string]T) []metricData {
	result := make([]metricData, 0, len(metrics))
	for name, value := range metrics {
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

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := pageData{
		Gauges:   mapMetrics(h.storage.GetAllGauges()),
		Counters: mapMetrics(h.storage.GetAllCounters()),
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
