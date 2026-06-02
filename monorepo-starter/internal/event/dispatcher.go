package event

import (
	"context"
	"fmt"
	"sync"
)

type Handler func(ctx context.Context, evt Envelope) (Envelope, error)

type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: make(map[string]Handler)}
}

func (d *Dispatcher) Register(eventType string, h Handler) {
	if eventType == "" || h == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = h
}

func (d *Dispatcher) Dispatch(ctx context.Context, evt Envelope) (Envelope, error) {
	d.mu.RLock()
	h, ok := d.handlers[evt.EventType]
	d.mu.RUnlock()
	if !ok {
		return Envelope{}, fmt.Errorf("no handler registered for event type %q", evt.EventType)
	}
	return h(ctx, evt)
}
