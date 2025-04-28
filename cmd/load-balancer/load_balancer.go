package main

import (
	"balancer/internal/balancer"
	"balancer/internal/config"
	"balancer/internal/proxy"
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
	conf, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}
	srvPool := balancer.NewServerPool(conf.Backends)
	algorithm := balancer.NewRoundRobin(srvPool)
	balancer := proxy.NewProxyHandler(algorithm)
	mux := http.NewServeMux()
	mux.HandleFunc("/", balancer.ServeHTTP)
	addr := ":" + strconv.Itoa(conf.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		log.Printf("starting balancer on port: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown: %v", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
