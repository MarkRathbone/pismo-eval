package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"event-processor/internal/delivery"
	"event-processor/internal/model"
	"event-processor/internal/storage"
	"event-processor/internal/validator"
)

func HandleMessage(payload string) error {
	// 1) Validate JSON schema
	if err := validator.Validate([]byte(payload)); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2) Unmarshal
	var evt model.Event
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 3) Save to DynamoDB
	if err := storage.SaveEvent(evt); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// 4) Optionally dispatch over HTTP
	if strings.ToLower(os.Getenv("DIRECT_DISPATCH")) == "true" {
		if err := delivery.DispatchEvent(evt); err != nil {
			return fmt.Errorf("failed to dispatch event: %w", err)
		}
	}

	return nil
}
