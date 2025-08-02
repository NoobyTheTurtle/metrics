// Package json предоставляет HTTP обработчики для JSON API метрик.
// Реализует REST эндпоинты для отправки и получения метрик в JSON формате.
package json

import "net/http"

type Handler struct {
	storage HandlerStorage
}

func NewHandler(storage HandlerStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// UpdateHandler возвращает HTTP обработчик для обновления отдельных метрик.
// Endpoint: POST /update/
func (h *Handler) UpdateHandler() http.HandlerFunc {
	handler := newUpdateHandler(h.storage)
	return handler.ServeHTTP
}

// ValueHandler возвращает HTTP обработчик для получения отдельных метрик.
// Endpoint: POST /value/
func (h *Handler) ValueHandler() http.HandlerFunc {
	handler := newValueHandler(h.storage)
	return handler.ServeHTTP
}

// UpdatesHandler возвращает HTTP обработчик для пакетного обновления метрик.
// Endpoint: POST /updates/
func (h *Handler) UpdatesHandler() http.HandlerFunc {
	handler := newUpdatesHandler(h.storage)
	return handler.ServeHTTP
}
