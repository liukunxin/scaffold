package model

// PingQuery is the incoming app-level query model.
type PingQuery struct {
	Name string
}

// PingResult is the app-level response model returned to transport layer.
type PingResult struct {
	EventID   string `json:"eventId"`
	EventType string `json:"eventType"`
	Seq       int64  `json:"seq"`
	Message   string `json:"message"`
}
