package ai

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"go-infra-starter/internal/infra/config"

	"github.com/liukunxin/go-infra/pkg/infra/llm"
)

var llmEnabled atomic.Bool

func InitLLM(cfg *config.App) error {
	if len(cfg.LLM.Providers) == 0 {
		llmEnabled.Store(false)
		return nil
	}
	if err := llm.InitFromConfig(cfg.LLM); err != nil {
		llmEnabled.Store(false)
		return err
	}
	llmEnabled.Store(true)
	return nil
}

func AskText(ctx context.Context, prompt string) (string, error) {
	if !llmEnabled.Load() {
		return "", errors.New("llm is not installed or not configured")
	}
	if strings.TrimSpace(prompt) == "" {
		prompt = "ping"
	}
	return llm.GetClient().AskText(ctx, prompt)
}
