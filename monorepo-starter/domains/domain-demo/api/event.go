package api

const EventPing = "domain.demo.ping"

type PingRequest struct {
	Name string `json:"name"`
}

type PingResponse struct {
	Message string `json:"message"`
}
