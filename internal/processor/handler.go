package processor

import (
	"context"
	"encoding/json"
	"log"

	"event-processor/internal/model"
	"event-processor/internal/storage"
	"event-processor/internal/validator"
)

func HandleMessage(ctx context.Context, payload string) error {
	if err := validator.Validate([]byte(payload)); err != nil {
		log.Println("Invalid event:", err)
		return err
	}

	var evt model.Event
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		log.Println("Unmarshal error:", err)
		return err
	}

	if err := storage.SaveEvent(ctx, evt); err != nil {
		log.Println("Storage error:", err)
		return err
	}

	return nil
}
