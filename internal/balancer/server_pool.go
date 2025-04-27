package balancer

import (
	"log"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	IsAlive      atomic.Bool
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
}

func NewServerPool(backendURLs []string) *ServerPool {
	var backends []*Backend

	for _, backendStr := range backendURLs {
		parsedURL, err := url.Parse(backendStr)
		if err != nil {
			log.Fatalf("Address parsing error %s: %v", backendStr, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(parsedURL)

		backend := &Backend{
			URL:          parsedURL,
			IsAlive:      atomic.Bool{},
			ReverseProxy: proxy,
		}
		backend.IsAlive.Store(true)
		backends = append(backends, backend)
	}

	return &ServerPool{
		backends: backends,
	}
}

func (sp *ServerPool) GetBackendsLen() int {
	return len(sp.backends)
}
