package runtime

import (
	"context"
	"fmt"

	"go-infra-monorepo-starter/internal/event"
)

type Router struct {
	dispatcher *event.Dispatcher
}

func NewRouter(dispatcher *event.Dispatcher) *Router {
	return &Router{dispatcher: dispatcher}
}

func (r *Router) Route(ctx context.Context, evt event.Envelope) (event.Envelope, error) {
	if r == nil || r.dispatcher == nil {
		return event.Envelope{}, fmt.Errorf("runtime router is not initialized")
	}
	return r.dispatcher.Dispatch(ctx, evt)
}
