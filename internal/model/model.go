package model

import "encoding/json"

type Event struct {
	ClientID  string          `json:"client_id"`
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}
