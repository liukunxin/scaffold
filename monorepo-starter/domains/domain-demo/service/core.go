package service

import (
	"context"
	"fmt"

	"go-infra-monorepo-starter/domains/domain-demo/api"
)

type Core struct{}

func NewCore() *Core {
	return &Core{}
}

func (c *Core) Ping(_ context.Context, req api.PingRequest) (api.PingResponse, error) {
	name := req.Name
	if name == "" {
		name = "world"
	}
	return api.PingResponse{Message: fmt.Sprintf("pong, %s", name)}, nil
}
