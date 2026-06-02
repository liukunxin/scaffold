package wsproto

import "go-infra-monorepo-starter/internal/event"

// Message is websocket transport protocol payload.
type Message struct {
	Type  string         `json:"type"`
	Event event.Envelope `json:"event"`
}
