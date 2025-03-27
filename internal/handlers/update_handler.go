package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *handler) updateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")

		if len(parts) != 3 {
			http.Error(w, "Invalid request format", http.StatusNotFound)
			return
		}

		metricType := MetricType(parts[0])
		metricName := parts[1]
		metricValue := parts[2]

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		switch metricType {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			err = h.storage.UpdateGauge(metricName, value)
			if err != nil {
				http.Error(w, "Failed to update gauge", http.StatusInternalServerError)
				return
			}
		case Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			err = h.storage.UpdateCounter(metricName, value)
			if err != nil {
				http.Error(w, "Failed to update counter", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
