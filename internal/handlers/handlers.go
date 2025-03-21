package handlers

import "net/http"

type handler struct {
	storage ServerStorage
}

func InitHandlers(storage ServerStorage) error {
	h := &handler{
		storage,
	}

	mux := http.NewServeMux()
	mux.Handle("/update/", setContentTypeMiddleware(h.updateHandler()))

	return http.ListenAndServe("localhost:8080", mux)
}
