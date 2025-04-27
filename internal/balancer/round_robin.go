package balancer

import (
	"log"
	"sync/atomic"
)

type RoundRobin struct {
	pool    *ServerPool
	current atomic.Uint64
}

func NewRoundRobin(pool *ServerPool) *RoundRobin {
	return &RoundRobin{pool: pool}
}

func (rr *RoundRobin) GetNextBackend() *Backend {
	total := rr.pool.GetBackendsLen()
	if total == 0 {
		return nil
	}

	startIdx := rr.NextIndex()
	for i := 0; i < total; i++ {
		idx := (int(startIdx) + i) % total
		backend := rr.pool.backends[idx]
		if backend.IsAlive.Load() {
			log.Printf("Selected backend: %s", backend.URL.String())
			return backend
		}
	}

	log.Println("No alive backend available")
	return nil
}

func (rr *RoundRobin) NextIndex() uint64 {
	total := uint64(rr.pool.GetBackendsLen())
	if total == 0 {
		return 0
	}
	return rr.current.Add(1) % total
}
