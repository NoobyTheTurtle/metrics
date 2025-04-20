package plain

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type valueGaugeHandler struct {
	storage GaugeGetter
}

type valueCounterHandler struct {
	storage CounterGetter
}

func newValueGaugeHandler(storage GaugeGetter) *valueGaugeHandler {
	return &valueGaugeHandler{
		storage: storage,
	}
}

func newValueCounterHandler(storage CounterGetter) *valueCounterHandler {
	return &valueCounterHandler{
		storage: storage,
	}
}

func (h *valueGaugeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	value, exists := h.storage.GetGauge(metricName)

	if !exists {
		http.Error(w, "Gauge not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
}

func (h *valueCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	value, exists := h.storage.GetCounter(metricName)

	if !exists {
		http.Error(w, "Counter not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatInt(value, 10))
}
