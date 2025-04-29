package ratelimiting

import (
	"balancer/internal/proxy"
	"log"
	"net/http"
)

var (
	RateLimitsError = "Rate limit exceeded"
)

// middleware checking clients for having at least one token
func (l *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем идентификатор клиента
		clientID := getClientIdentifier(r)

		// checking limits
		ok, err := l.Allow(r.Context(), clientID)
		if !ok {
			if err != nil {
				log.Printf("[ERROR] Checking token balance: %s %v", clientID, err)
			} else {
				log.Printf("[ERROR] Too many requests: %s", clientID)
			}
			proxy.RespondWithError(w, http.StatusTooManyRequests, RateLimitsError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIdentifier(r *http.Request) string {
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "api_key:" + apiKey
	}

	return "ip:" + r.RemoteAddr
}
