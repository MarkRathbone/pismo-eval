package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"event-processor/internal/delivery"
)

func main() {
	log.Println("Starting delivery service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutdown signal received, canceling delivery processor...")
		cancel()
	}()

	if err := delivery.StartStreamProcessor(ctx); err != nil {
		log.Fatalf("Stream processor failed: %v", err)
	}

	log.Println("Delivery service shut down cleanly.")
}
