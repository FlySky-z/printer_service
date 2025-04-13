package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"printer/config"
	"printer/router"
	"syscall"
	"time"
)

func interruptServer(server *http.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		log.Println("Shutting down server gracefully...")

		// Create a context with timeout for graceful shutdown
		timeout := 5 // seconds
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		} else {
			log.Println("Server shut down gracefully")
		}
	}()
}
func main() {
	config.SetGinMode("release")
	r := router.SetupRouter()

	server := &http.Server{
		Addr:    "0.0.0.0:80",
		Handler: r,
	}

	interruptServer(server)
	// 当前监听的地址
	log.Printf("Server is listening on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
