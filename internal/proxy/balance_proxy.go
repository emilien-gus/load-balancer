package proxy

import (
	"balancer/internal/balancer"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ProxyHandler struct {
	balancer balancer.Balancer
}

func NewProxyHandler(bal balancer.Balancer) *ProxyHandler {
	return &ProxyHandler{
		balancer: bal,
	}
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	backend := p.balancer.GetNextBackend()
	if backend == nil {
		log.Printf("[ERROR] No available backend to handle request: %s %s", r.Method, r.URL.Path)
		RespondWithError(w, http.StatusServiceUnavailable, "Service unavailable")
		return
	}

	log.Printf("[INFO] Forwarding request %s %s to backend %s", r.Method, r.URL.Path, backend.URL)

	backend.ReverseProxy.ServeHTTP(w, r)

	duration := time.Since(start)
	log.Printf("[INFO] Request %s %s completed in %v", r.Method, r.URL.Path, duration)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	_ = json.NewEncoder(w).Encode(response)
}
