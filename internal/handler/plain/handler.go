package plain

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage HandlerStorage
}

func NewHandler(storage HandlerStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) ValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := MetricType(chi.URLParam(r, "metricType"))

		switch metricType {
		case Gauge:
			handler := newValueGaugeHandler(h.storage)
			handler.ServeHTTP(w, r)
		case Counter:
			handler := newValueCounterHandler(h.storage)
			handler.ServeHTTP(w, r)
		default:
			http.Error(w, "Unknown metric type", http.StatusNotFound)
		}
	}
}

func (h *Handler) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := MetricType(chi.URLParam(r, "metricType"))

		switch metricType {
		case Gauge:
			handler := newUpdateGaugeHandler(h.storage)
			handler.ServeHTTP(w, r)
		case Counter:
			handler := newUpdateCounterHandler(h.storage)
			handler.ServeHTTP(w, r)
		default:
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
		}
	}
}
