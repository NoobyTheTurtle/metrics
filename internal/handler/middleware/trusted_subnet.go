package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware создает middleware для проверки доверенной подсети
func TrustedSubnetMiddleware(trustedSubnet string, logger MiddlewareLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				logger.Error("Invalid CIDR format for trusted subnet '%s': %v", trustedSubnet, err)
				http.Error(w, "Server configuration error", http.StatusInternalServerError)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				logger.Info("Request without X-Real-IP header from %s for %s", r.RemoteAddr, r.URL.Path)
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				logger.Info("Invalid IP address in X-Real-IP header '%s' from %s for %s", realIP, r.RemoteAddr, r.URL.Path)
				http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusBadRequest)
				return
			}

			if !subnet.Contains(clientIP) {
				logger.Info("Request from untrusted IP %s (via %s) for %s - not in trusted subnet %s", realIP, r.RemoteAddr, r.URL.Path, trustedSubnet)
				http.Error(w, "Access denied: IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
