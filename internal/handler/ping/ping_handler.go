package ping

import (
	"net/http"
)

func (h *Handler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.db.Ping(r.Context())
		if err != nil {
			h.logger.Error("Database connection failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
