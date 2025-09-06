package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// DecryptMiddleware создает middleware для дешифрования тела запроса, если доступен дешифратор
func DecryptMiddleware(decrypter Decrypter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if decrypter == nil {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "Failed to read request body")
				return
			}

			decryptedData, err := decrypter.Decrypt(body)
			if err != nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
			} else {
				r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			}

			next.ServeHTTP(w, r)
		})
	}
}
