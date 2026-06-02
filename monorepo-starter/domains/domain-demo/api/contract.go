package api

import "context"

type Contract interface {
	Ping(ctx context.Context, req PingRequest) (PingResponse, error)
}
