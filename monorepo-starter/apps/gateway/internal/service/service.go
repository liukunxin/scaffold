package service

import (
	"context"
	"fmt"
	"strings"

	"go-infra-monorepo-starter/apps/gateway/internal/model"
	"go-infra-monorepo-starter/apps/gateway/internal/repository"
	demoapi "go-infra-monorepo-starter/domains/domain-demo/api"
	runtimecore "go-infra-monorepo-starter/internal/runtime"
)

// RuntimeService orchestrates app input -> runtime event -> app output.
type RuntimeService struct {
	engine *runtimecore.Engine
	repo   repository.ConfigRepository
}

func NewRuntimeService(engine *runtimecore.Engine, repo repository.ConfigRepository) *RuntimeService {
	return &RuntimeService{
		engine: engine,
		repo:   repo,
	}
}

func (s *RuntimeService) Ping(ctx context.Context, in model.PingQuery) (model.PingResult, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = s.repo.DefaultPingName()
	}

	out, err := s.engine.Handle(ctx, demoapi.EventPing, map[string]any{
		"name": name,
	})
	if err != nil {
		return model.PingResult{}, fmt.Errorf("runtime ping failed: %w", err)
	}

	message, _ := out.Payload["message"].(string)
	return model.PingResult{
		EventID:   out.EventID,
		EventType: out.EventType,
		Seq:       out.Seq,
		Message:   message,
	}, nil
}
