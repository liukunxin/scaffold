package controller

import (
	"io"

	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"go-infra-starter/internal/app/llm/dto"
	"go-infra-starter/internal/app/llm/logic"
	"go-infra-starter/internal/app/llm/ro"
)

type LLMController interface {
	// Ping 验证LLM链路可用性并返回模型回复。
	Ping(ctx *gin.Context)
}

type llmController struct {
	kcontroller.GinBase
	llmLogic logic.LLMLogic
}

func NewLLMController(llmLogic logic.LLMLogic) LLMController {
	return &llmController{llmLogic: llmLogic}
}

func (c *llmController) Ping(ctx *gin.Context) {
	var req ro.PingReq
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.ErrorResponse(ctx, err)
		return
	}

	data, err := c.llmLogic.Ping(ctx, &dto.PingInput{Prompt: req.Prompt})
	if err != nil {
		c.ErrorResponse(ctx, err)
		return
	}
	c.SuccessResponse(ctx, data)
}
