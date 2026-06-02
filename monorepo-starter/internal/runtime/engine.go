package runtime

import (
	"context"
	"fmt"
	"time"

	"go-infra-monorepo-starter/internal/event"
)

type Engine struct {
	seq    *Sequencer
	router *Router
}

func NewEngine(router *Router) *Engine {
	return &Engine{
		seq:    NewSequencer(),
		router: router,
	}
}

func (e *Engine) Handle(ctx context.Context, eventType string, payload map[string]any) (event.Envelope, error) {
	if e == nil || e.router == nil {
		return event.Envelope{}, fmt.Errorf("runtime engine is not initialized")
	}
	if eventType == "" {
		return event.Envelope{}, fmt.Errorf("event type is required")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	seq := e.seq.Next()
	req := event.Envelope{
		EventID:   fmt.Sprintf("evt-%d", seq),
		EventType: eventType,
		Seq:       seq,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}
	return e.router.Route(ctx, req)
}
