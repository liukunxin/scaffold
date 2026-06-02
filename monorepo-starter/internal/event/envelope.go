package event

// Envelope is the canonical event envelope across runtime and domains.
type Envelope struct {
	EventID   string         `json:"eventId"`
	EventType string         `json:"eventType"`
	SessionID string         `json:"sessionId"`
	Seq       int64          `json:"seq"`
	Timestamp int64          `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}
