package ping

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err := h.db.Ping(ctx)
		if err != nil {
			h.logger.Error("Database connection failed: %v", err)
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
