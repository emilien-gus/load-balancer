package main

import (
	"balancer/internal/balancer"
	"balancer/internal/config"
	"balancer/internal/data"
	"balancer/internal/proxy"
	"balancer/internal/ratelimiting"
	"balancer/internal/repository"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// init connection with database
	if err := data.InitDB(); err != nil {
		log.Fatalf("failed to open connection with db: %v", err)
	}
	defer func() {
		if err := data.CloseDB(); err != nil {
			log.Printf("error at closing connection with db: %v", err)
		}
		log.Printf("connection with db is closed")
	}()

	// getting config from file
	conf, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	if conf.BucketCapacity <= 0 || conf.RatePerSec < 0 {
		log.Fatal("wrong config parametrs")
	}

	// init balancer logic
	srvPool := balancer.NewServerPool(conf.Backends)
	algorithm := balancer.NewRoundRobin(srvPool)
	balancer := proxy.NewProxyHandler(algorithm)

	// init rate limit logic
	clientRepo := repository.NewClientsRepo(data.DB)
	rateLimit := ratelimiting.NewLimiter(conf.BucketCapacity, conf.RatePerSec, clientRepo)
	log.Printf("rate limiting configured: capacity=%d, rate=%d/sec", conf.BucketCapacity, conf.RatePerSec)
	defer rateLimit.Stop()

	// setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", balancer.ServeHTTP)
	limitMux := rateLimit.RateLimitMiddleware(mux)

	addr := ":" + strconv.Itoa(conf.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: limitMux,
	}

	go func() {
		log.Printf("starting balancer on port: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// waiting for SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown: %v", err)
	}

	select {
	case <-ctx.Done():
		log.Println("Shutdown completed after timeout.")
	default:
		log.Println("Server exited gracefully.")
	}

	log.Println("Server exiting")
}
