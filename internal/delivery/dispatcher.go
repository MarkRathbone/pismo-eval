package delivery

import (
	"bytes"
	"encoding/json"
	"event-processor/internal/model"
	"fmt"
	"log"
	"net/http"
)

var clientRoutes = map[string]string{
	"client-123": "http://localhost:8081/events",
}

func DispatchEvent(event model.Event) {
	target, ok := clientRoutes[event.ClientID]
	if !ok {
		log.Printf("No route for client_id: %s", event.ClientID)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	resp, err := http.Post(target, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("Failed to deliver event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Delivery failed with status: %d", resp.StatusCode)
		return
	}

	log.Printf("Event delivered to %s successfully", target)
}
