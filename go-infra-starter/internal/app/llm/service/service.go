package service

import (
	"context"

	"go-infra-starter/internal/infra/ai"
)

type LLMService interface {
	// AskText 调用LLM文本问答能力。
	AskText(ctx context.Context, prompt string) (string, error)
}

type llmService struct{}

func NewLLMService() LLMService {
	return &llmService{}
}

func (s *llmService) AskText(ctx context.Context, prompt string) (string, error) {
	return ai.AskText(ctx, prompt)
}
