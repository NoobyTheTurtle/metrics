package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type valueGaugeHandler struct {
	storage gaugeGetter
	logger  Logger
}

type valueCounterHandler struct {
	storage counterGetter
	logger  Logger
}

func newValueGaugeHandler(storage gaugeGetter, logger Logger) *valueGaugeHandler {
	return &valueGaugeHandler{
		storage: storage,
		logger:  logger,
	}
}

func newValueCounterHandler(storage counterGetter, logger Logger) *valueCounterHandler {
	return &valueCounterHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *valueGaugeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	metricName := chi.URLParam(r, "metricName")
	value, exists := h.storage.GetGauge(metricName)

	if !exists {
		http.Error(w, "Gauge not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
}

func (h *valueCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	metricName := chi.URLParam(r, "metricName")
	value, exists := h.storage.GetCounter(metricName)

	if !exists {
		http.Error(w, "Counter not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatInt(value, 10))
}

func (h *handler) valueHandler() http.HandlerFunc {
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
