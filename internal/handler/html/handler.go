package html

import (
	"net/http"
)

type Handler struct {
	storage HandlerStorage
}

func NewHandler(storage HandlerStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) IndexHandler() http.HandlerFunc {
	handler := newIndexHandler(h.storage)
	return handler.ServeHTTP
}
