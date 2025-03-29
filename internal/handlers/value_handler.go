package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *handler) valueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")

		metricType := MetricType(chi.URLParam(r, "metricType"))
		metricName := chi.URLParam(r, "metricName")

		switch metricType {
		case Gauge:
			value, exists := h.storage.GetGauge(metricName)

			if !exists {
				http.Error(w, "Gauge not found", http.StatusNotFound)
				return
			}

			io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))

		case Counter:
			value, exists := h.storage.GetCounter(metricName)

			if !exists {
				http.Error(w, "Counter not found", http.StatusNotFound)
				return
			}

			io.WriteString(w, strconv.FormatInt(value, 10))
		default:
			http.Error(w, "Unknown metric type", http.StatusNotFound)
		}
	}
}
