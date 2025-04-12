package plain

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage HandlerStorage
	logger  HandlerLogger
}

func NewHandler(storage HandlerStorage, logger HandlerLogger) *Handler {
	return &Handler{
		storage: storage,
		logger:  logger,
	}
}

func (h *Handler) IndexHandler() http.HandlerFunc {
	handler := newIndexHandler(h.storage, h.logger)
	return handler.ServeHTTP
}

func (h *Handler) ValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := MetricType(chi.URLParam(r, "metricType"))

		switch metricType {
		case Gauge:
			handler := newValueGaugeHandler(h.storage, h.logger)
			handler.ServeHTTP(w, r)
		case Counter:
			handler := newValueCounterHandler(h.storage, h.logger)
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
			handler := newUpdateGaugeHandler(h.storage, h.logger)
			handler.ServeHTTP(w, r)
		case Counter:
			handler := newUpdateCounterHandler(h.storage, h.logger)
			handler.ServeHTTP(w, r)
		default:
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
		}
	}
}
