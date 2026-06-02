package runtime

import (
	"context"
	"fmt"

	"go-infra-monorepo-starter/domains/domain-demo/api"
	"go-infra-monorepo-starter/internal/event"
)

func NewHandler(contract api.Contract) func(ctx context.Context, evt event.Envelope) (event.Envelope, error) {
	if contract == nil {
		panic("domain-demo runtime handler requires non-nil contract")
	}
	return func(ctx context.Context, evt event.Envelope) (event.Envelope, error) {
		req := api.PingRequest{}
		if v, ok := evt.Payload["name"].(string); ok {
			req.Name = v
		}
		resp, err := contract.Ping(ctx, req)
		if err != nil {
			return event.Envelope{}, fmt.Errorf("domain-demo ping failed: %w", err)
		}
		return event.Envelope{
			EventID:   evt.EventID,
			EventType: api.EventPing + ".result",
			SessionID: evt.SessionID,
			Seq:       evt.Seq,
			Timestamp: evt.Timestamp,
			Payload: map[string]any{
				"message": resp.Message,
			},
		}, nil
	}
}
