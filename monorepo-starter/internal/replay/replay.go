package replay

import (
	"context"

	"go-infra-monorepo-starter/internal/event"
)

// Runner replays an ordered event stream through runtime.
type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Replay(ctx context.Context, events []event.Envelope, apply func(context.Context, event.Envelope) error) error {
	for _, evt := range events {
		if err := apply(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
