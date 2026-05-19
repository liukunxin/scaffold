package controller

import (
	"io"

	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"go-infra-starter/internal/app/llm/dto"
	"go-infra-starter/internal/app/llm/ro"
	"go-infra-starter/internal/app/llm/service"
	"go-infra-starter/internal/app/llm/vo"
)

type LLMController interface {
	// Ping 验证LLM链路可用性并返回模型回复。
	Ping(ctx *gin.Context)
}

type llmController struct {
	kcontroller.GinBase
	llmService service.LLMService
}

func NewLLMController(llmService service.LLMService) LLMController {
	return &llmController{llmService: llmService}
}

func (c *llmController) Ping(ctx *gin.Context) {
	var req ro.PingReq
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.ErrorResponse(ctx, err)
		return
	}

	data, err := c.llmService.AskText(ctx, dto.PingInput{Prompt: req.Prompt})
	if err != nil {
		c.ErrorResponse(ctx, err)
		return
	}
	c.SuccessResponse(ctx, &vo.PingResp{Reply: data})
}
