package logic

import (
	"context"

	"go-infra-starter/internal/app/llm/dto"
	"go-infra-starter/internal/app/llm/service"
	"go-infra-starter/internal/app/llm/vo"
	"go-infra-starter/internal/infra/ai"
)

type LLMLogic interface {
	Ping(ctx context.Context, in *dto.PingInput) (*vo.PingResp, error)
}

type llmLogic struct {
	llmSvc service.LLMService
}

func NewLLMLogic(llmSvc service.LLMService) LLMLogic {
	return &llmLogic{llmSvc: llmSvc}
}

func (l *llmLogic) Ping(ctx context.Context, in *dto.PingInput) (*vo.PingResp, error) {
	reply, err := l.llmSvc.AskText(ctx, in.Prompt)
	if err != nil {
		return nil, err
	}
	return &vo.PingResp{
		Enabled: ai.Enabled(),
		Reply:   reply,
	}, nil
}
