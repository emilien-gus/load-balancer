package ratelimiting

import (
	"sync"
	"sync/atomic"
)

type TokenBucket struct {
	capacity   int32
	tokens     atomic.Int32
	ratePerSec int
	mutex      sync.RWMutex
}

func NewTokenBucket(capacity, ratePerSec int) *TokenBucket {
	backet := &TokenBucket{
		capacity:   int32(capacity),
		tokens:     atomic.Int32{},
		ratePerSec: ratePerSec,
		mutex:      sync.RWMutex{},
	}
	backet.tokens.Store(int32(capacity))
	return backet
}

func (tb *TokenBucket) SetCapacity(newCapacity int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.capacity = int32(newCapacity)
}

func (tb *TokenBucket) SetRatePerSecond(newRatePerSec int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.ratePerSec = newRatePerSec
}

func (tb *TokenBucket) refillBucket() {
	tb.mutex.RLock()
	rate := tb.ratePerSec
	cap := tb.capacity
	tb.mutex.RUnlock()

	currentTokens := tb.tokens.Load()
	newTokensVal := currentTokens + int32(rate)

	if newTokensVal > cap {
		newTokensVal = tb.capacity
	}

	tb.tokens.Store(newTokensVal)
}

func (tb *TokenBucket) checkNotEmpty() bool {
	currentTokens := tb.tokens.Load()
	if currentTokens <= 0 {
		return false
	}
	tb.tokens.Store(currentTokens - 1)
	return true
}
