package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	storage ServerStorage
}

func InitHandlers(serverAddress string, storage ServerStorage) error {
	h := &handler{
		storage,
	}

	r := chi.NewRouter()
	r.Use(loggingMiddleware)
	r.Get("/", h.indexHandler())
	r.Get("/value/{metricType}/{metricName}", h.valueHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.updateHandler())

	return http.ListenAndServe(serverAddress, r)
}
