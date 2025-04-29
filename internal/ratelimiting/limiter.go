package ratelimiting

import (
	"balancer/internal/repository"
	"context"
	"fmt"
	"sync"
	"time"
)

var tickerInterval = time.Second

type RateLimiter struct {
	buckets     sync.Map
	defaultCap  int32
	defaultRate int
	stopChan    chan struct{}
	storage     repository.ClientsRepoInterface
}

func NewLimiter(defaultCap int32, defaultRate int, storage repository.ClientsRepoInterface) *RateLimiter {
	limiter := &RateLimiter{
		buckets:     sync.Map{},
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
		stopChan:    make(chan struct{}, 1),
		storage:     storage,
	}

	go limiter.refillBuckets()
	return limiter
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// 1. Проверяем существующий бакет
	if val, ok := rl.buckets.Load(key); ok {
		bucket := val.(*TokenBucket)
		if bucket.checkNotEmpty() {
			return true, nil
		}
		return false, nil
	}

	// 2. Создаем новый бакет
	capacity, rate, err := rl.storage.GetOrCreate(ctx, key, rl.defaultCap, rl.defaultRate)
	if err != nil {
		return false, fmt.Errorf("failed to get/create bucket: %w", err)
	}

	// 3. Валидация параметров
	if capacity <= 0 || rate <= 0 {
		return false, fmt.Errorf("invalid bucket parameters: capacity=%d, rate=%d", capacity, rate)
	}

	bucket := NewTokenBucket(int(capacity), rate)
	rl.buckets.Store(key, bucket)

	return bucket.checkNotEmpty(), nil
}

func (rl *RateLimiter) refillBuckets() {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.buckets.Range(func(key, value any) bool {
				bucket := value.(*TokenBucket)
				bucket.refillBucket()
				return true
			})
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}
