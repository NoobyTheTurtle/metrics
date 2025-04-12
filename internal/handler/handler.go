package handler

import (
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/handler/middleware"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func InitHandlers(serverAddress string, store *storage.MemStorage, log *logger.ZapLogger) error {
	plainHandler := plain.NewHandler(store, log)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(log))

	// Plain handlers
	r.Get("/", plainHandler.IndexHandler())
	r.Get("/value/{metricType}/{metricName}", plainHandler.ValueHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", plainHandler.UpdateHandler())

	log.Info("Starting server on %s", serverAddress)
	return http.ListenAndServe(serverAddress, r)
}
