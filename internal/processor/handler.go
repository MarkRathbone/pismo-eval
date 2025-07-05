package processor

import (
	"encoding/json"
	"event-processor/internal/delivery"
	"event-processor/internal/model"
	"event-processor/internal/storage"
	"event-processor/internal/validator"
	"log"
)

func HandleMessage(payload string) {
	if err := validator.Validate([]byte(payload)); err != nil {
		log.Println("Invalid event:", err)
		return
	}

	var evt model.Event
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	if err := storage.SaveEvent(evt); err != nil {
		log.Println("Storage error:", err)
		return
	}

	delivery.DispatchEvent(evt)
}
