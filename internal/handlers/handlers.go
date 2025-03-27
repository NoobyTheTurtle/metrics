package handlers

import (
	"net/http"
)

type handler struct {
	storage ServerStorage
}

func InitHandlers(serverAddress string, storage ServerStorage) error {
	h := &handler{
		storage,
	}

	mux := http.NewServeMux()
	mux.Handle("/update/", conveyor(h.updateHandler(), loggingMiddleware))

	return http.ListenAndServe(serverAddress, mux)
}
