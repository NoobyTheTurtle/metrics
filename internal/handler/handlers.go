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
	r.Use(middleware.LogMiddleware(log))

	// HTML handlers
	htmlHandler := html.NewHandler(store)
	htmlRouter := chi.NewRouter()
	htmlRouter.Use(middleware.ContentTypeMiddleware(html.ContentTypeValue))
	htmlRouter.Get("/", htmlHandler.IndexHandler())
	r.Mount("/", htmlRouter)

	// Plain handlers
	plainHandler := plain.NewHandler(store)
	plainRouter := chi.NewRouter()
	plainRouter.Use(middleware.ContentTypeMiddleware(plain.ContentTypeValue))
	plainRouter.Get("/value/{metricType}/{metricName}", plainHandler.ValueHandler())
	plainRouter.Post("/update/{metricType}/{metricName}/{metricValue}", plainHandler.UpdateHandler())
	r.Mount("/", plainRouter)

	// JSON handlers
	jsonHandler := json.NewHandler(store)
	jsonRouter := chi.NewRouter()
	jsonRouter.Use(middleware.ContentTypeMiddleware(json.ContentTypeValue))
	jsonRouter.Post("/update", jsonHandler.UpdateHandler())
	jsonRouter.Post("/value", jsonHandler.ValueHandler())
	r.Mount("/", jsonRouter)

	log.Info("Starting server on %s", serverAddress)
	return http.ListenAndServe(serverAddress, r)
}
