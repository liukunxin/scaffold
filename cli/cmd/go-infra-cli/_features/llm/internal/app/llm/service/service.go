package service

import (
	"context"

	"single-starter/internal/app/llm/dto"
	"single-starter/internal/infra/ai"
)

type LLMService interface {
	// AskText 调用LLM文本问答能力。
	AskText(ctx context.Context, in dto.PingInput) (string, error)
}

type llmService struct{}

func NewLLMService() LLMService {
	return &llmService{}
}

func (s *llmService) AskText(ctx context.Context, in dto.PingInput) (string, error) {
	return ai.AskText(ctx, in.Prompt)
}
