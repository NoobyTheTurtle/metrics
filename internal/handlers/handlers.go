package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	storage serverStorage
	logger  handlersLogger
}

func InitHandlers(serverAddress string, storage serverStorage, log handlersLogger) error {
	h := &handler{
		storage: storage,
		logger:  log,
	}

	r := chi.NewRouter()
	r.Use(loggingMiddleware(log))
	r.Get("/", h.indexHandler())
	r.Get("/value/{metricType}/{metricName}", h.valueHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.updateHandler())

	log.Info("Starting server on %s", serverAddress)
	return http.ListenAndServe(serverAddress, r)
}
