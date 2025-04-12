package json

import (
	"io"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type UpdateStorage interface {
	GaugeSetter
	CounterSetter
}

type updateHandler struct {
	storage UpdateStorage
}

func newUpdateHandler(storage UpdateStorage) *updateHandler {
	return &updateHandler{
		storage: storage,
	}
}

func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		if metric.Value == nil {
			http.Error(w, "Value field is required for gauge type", http.StatusBadRequest)
			return
		}

		value, err := h.storage.UpdateGauge(metric.ID, *metric.Value)
		if err != nil {
			http.Error(w, "Failed to update gauge", http.StatusInternalServerError)
			return
		}

		metric.Value = &value
	case CounterType:
		if metric.Delta == nil {
			http.Error(w, "Delta field is required for counter type", http.StatusBadRequest)
			return
		}

		value, err := h.storage.UpdateCounter(metric.ID, *metric.Delta)
		if err != nil {
			http.Error(w, "Failed to update counter", http.StatusInternalServerError)
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
