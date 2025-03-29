package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	storage ServerStorage
	logger  Logger
}

func InitHandlers(serverAddress string, storage ServerStorage, log Logger) error {
	h := &handler{
		storage: storage,
		logger:  log,
	}

	r := chi.NewRouter()
	r.Use(loggingMiddleware(log))
	r.Get("/", h.indexHandler())
	r.Get("/value/{metricType}/{metricName}", h.valueHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.updateHandler())

	return http.ListenAndServe(serverAddress, r)
}
