package service

import (
	"context"
	"fmt"

	"go-infra-starter/internal/app/demo/dto"
	"go-infra-starter/internal/app/demo/vo"
)

type DemoService interface {
	// Ping 返回演示用 pong 消息。
	Ping(ctx context.Context, in dto.PingInput) (*vo.PingResp, error)
}

type demoService struct{}

func NewDemoService() DemoService {
	return &demoService{}
}

func (s *demoService) Ping(_ context.Context, in dto.PingInput) (*vo.PingResp, error) {
	name := in.Name
	if name == "" {
		name = "world"
	}
	return &vo.PingResp{Message: fmt.Sprintf("pong, %s", name)}, nil
}
