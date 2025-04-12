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

func (h *Handler) ValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement
	}
}

func (h *Handler) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement
	}
}
