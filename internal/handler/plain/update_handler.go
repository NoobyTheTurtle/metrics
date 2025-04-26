package plain

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type updateGaugeHandler struct {
	storage GaugeSetter
}

type updateCounterHandler struct {
	storage CounterSetter
}

func newUpdateGaugeHandler(storage GaugeSetter) *updateGaugeHandler {
	return &updateGaugeHandler{
		storage: storage,
	}
}

func newUpdateCounterHandler(storage CounterSetter) *updateCounterHandler {
	return &updateCounterHandler{
		storage: storage,
	}
}

func (h *updateGaugeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.storage.UpdateGauge(r.Context(), metricName, value)
	if err != nil {
		http.Error(w, "Failed to update gauge", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *updateCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.storage.UpdateCounter(r.Context(), metricName, value)
	if err != nil {
		http.Error(w, "Failed to update counter", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
