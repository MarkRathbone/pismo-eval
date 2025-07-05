package main

import (
	"event-processor/internal/delivery"
	"log"
)

func main() {
	log.Println("Starting delivery service...")

	err := delivery.StartStreamProcessor()
	if err != nil {
		log.Fatalf("Stream processor failed: %v", err)
	}
}
