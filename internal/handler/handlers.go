package handler

import (
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/handler/html"
	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/handler/middleware"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func InitHandlers(serverAddress string, store *storage.MemStorage, log *logger.ZapLogger) error {

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(log))

	// HTML handlers
	htmlHandler := html.NewHandler(store)
	r.Get("/", htmlHandler.IndexHandler())

	// Plain handlers
	plainHandler := plain.NewHandler(store)
	r.Get("/value/{metricType}/{metricName}", plainHandler.ValueHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", plainHandler.UpdateHandler())

	// JSON handlers
	jsonHandler := json.NewHandler(store)
	r.Post("/update", jsonHandler.UpdateHandler())
	r.Post("/value", jsonHandler.ValueHandler())

	log.Info("Starting server on %s", serverAddress)
	return http.ListenAndServe(serverAddress, r)
}
