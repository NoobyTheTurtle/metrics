package handler

import (
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/handler/html"
	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/handler/middleware"
	"github.com/NoobyTheTurtle/metrics/internal/handler/ping"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	router       chi.Router
	storage      MetricStorage
	logger       RouterLogger
	pingHandler  *ping.Handler
	htmlHandler  *html.Handler
	plainHandler *plain.Handler
	jsonHandler  *json.Handler
}

func NewRouter(storage MetricStorage, logger RouterLogger, dbClient DBPinger) *Router {
	r := &Router{
		router:  chi.NewRouter(),
		storage: storage,
		logger:  logger,
	}

	r.htmlHandler = html.NewHandler(storage)
	r.plainHandler = plain.NewHandler(storage)
	r.jsonHandler = json.NewHandler(storage)
	r.pingHandler = ping.NewHandler(dbClient, logger)
	r.setupMiddleware()
	r.setupRoutes()

	return r
}

func (r *Router) setupMiddleware() {
	r.router.Use(middleware.LogMiddleware(r.logger))
}

func (r *Router) setupRoutes() {
	// Ping handler
	r.router.Get("/ping", r.pingHandler.PingHandler())

	// HTML handlers
	r.router.Group(func(router chi.Router) {
		router.Use(middleware.ContentTypeMiddleware(html.ContentTypeValue))
		router.Use(middleware.GzipMiddleware)
		router.Get("/", r.htmlHandler.IndexHandler())
	})

	// Plain handlers
	r.router.Group(func(router chi.Router) {
		router.Use(middleware.ContentTypeMiddleware(plain.ContentTypeValue))
		router.Get("/value/{metricType}/{metricName}", r.plainHandler.ValueHandler())
		router.Post("/update/{metricType}/{metricName}/{metricValue}", r.plainHandler.UpdateHandler())
	})

	// JSON handlers
	r.router.Group(func(router chi.Router) {
		router.Use(middleware.ContentTypeMiddleware(json.ContentTypeValue))
		router.Use(middleware.GzipMiddleware)
		router.Post("/update/", r.jsonHandler.UpdateHandler())
		router.Post("/value/", r.jsonHandler.ValueHandler())
	})
}

func (r *Router) Handler() http.Handler {
	return r.router
}
