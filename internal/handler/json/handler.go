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

func (h *Handler) UpdateHandler() http.HandlerFunc {
	handler := newUpdateHandler(h.storage)
	return handler.ServeHTTP
}

func (h *Handler) ValueHandler() http.HandlerFunc {
	handler := newValueHandler(h.storage)
	return handler.ServeHTTP
}

func (h *Handler) UpdatesHandler() http.HandlerFunc {
	handler := newUpdatesHandler(h.storage)
	return handler.ServeHTTP
}
