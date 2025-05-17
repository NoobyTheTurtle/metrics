package json

import (
	"io"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type UpdatesStorage interface {
	GaugeSetter
	CounterSetter
	BatchUpdater
}

type updatesHandler struct {
	storage UpdatesStorage
}

func newUpdatesHandler(storage UpdatesStorage) *updatesHandler {
	return &updatesHandler{
		storage: storage,
	}
}

func (h *updatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metrics model.Metrics
	if err := metrics.UnmarshalJSON(body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	for _, metric := range metrics {
		if metric.ID == "" || metric.MType == "" {
			http.Error(w, "id and type fields are required for all metrics", http.StatusBadRequest)
			return
		}

		switch metric.MType {
		case model.GaugeType:
			if metric.Value == nil {
				http.Error(w, "value field is required for gauge type", http.StatusBadRequest)
				return
			}
		case model.CounterType:
			if metric.Delta == nil {
				http.Error(w, "delta field is required for counter type", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
			return
		}
	}

	err = h.storage.UpdateMetricsBatch(r.Context(), metrics)
	if err != nil {
		http.Error(w, "Failed to update metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
