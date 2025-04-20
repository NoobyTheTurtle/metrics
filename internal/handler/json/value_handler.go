package json

import (
	"io"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type ValueStorage interface {
	GaugeGetter
	CounterGetter
}

type valueHandler struct {
	storage ValueStorage
}

func newValueHandler(storage ValueStorage) *valueHandler {
	return &valueHandler{
		storage: storage,
	}
}

func (h *valueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := metric.UnmarshalJSON(body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if metric.ID == "" || metric.MType == "" {
		http.Error(w, "id and type fields are required", http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case GaugeType:
		value, exists := h.storage.GetGauge(metric.ID)
		if !exists {
			http.Error(w, "Gauge not found", http.StatusNotFound)
			return
		}

		metric.Value = &value
	case CounterType:
		value, exists := h.storage.GetCounter(metric.ID)
		if !exists {
			http.Error(w, "Counter not found", http.StatusNotFound)
			return
		}

		metric.Delta = &value
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	resp, err := metric.MarshalJSON()
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
