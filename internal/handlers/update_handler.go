package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type updateGaugeHandler struct {
	storage GaugeSetter
	logger  Logger
}

type updateCounterHandler struct {
	storage CounterSetter
	logger  Logger
}

func newUpdateGaugeHandler(storage GaugeSetter, logger Logger) *updateGaugeHandler {
	return &updateGaugeHandler{
		storage: storage,
		logger:  logger,
	}
}

func newUpdateCounterHandler(storage CounterSetter, logger Logger) *updateCounterHandler {
	return &updateCounterHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *updateGaugeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

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

	w.WriteHeader(http.StatusOK)
}

func (h *updateCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

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

	w.WriteHeader(http.StatusOK)
}

func (h *handler) updateHandler() http.HandlerFunc {
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
