package ai

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"go-infra-starter/internal/infra/config"

	"github.com/liukunxin/go-infra/pkg/infra/llm"
)

type LLMState struct {
	Enabled bool
}

var llmEnabled atomic.Bool

func InitLLM(cfg *config.App) (*LLMState, error) {
	if !cfg.Features.LLM {
		llmEnabled.Store(false)
		return &LLMState{Enabled: false}, nil
	}
	if len(cfg.LLM.Providers) == 0 {
		llmEnabled.Store(false)
		return &LLMState{Enabled: false}, nil
	}
	if err := llm.InitFromConfig(cfg.LLM); err != nil {
		llmEnabled.Store(false)
		return nil, err
	}
	llmEnabled.Store(true)
	return &LLMState{Enabled: true}, nil
}

func Enabled() bool {
	return llmEnabled.Load()
}

func AskText(ctx context.Context, prompt string) (string, error) {
	if !Enabled() {
		return "", errors.New("llm feature is disabled or not configured")
	}
	if strings.TrimSpace(prompt) == "" {
		prompt = "ping"
	}
	return llm.GetClient().AskText(ctx, prompt)
}
