package controller

import (
	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"go-infra-starter/internal/app/demo/dto"
	"go-infra-starter/internal/app/demo/service"
)

type DemoController interface {
	// Ping 三层架构示例：controller -> service（无 logic）。
	Ping(ctx *gin.Context)
}

type demoController struct {
	kcontroller.GinBase
	demoService service.DemoService
}

func NewDemoController(demoService service.DemoService) DemoController {
	return &demoController{demoService: demoService}
}

func (c *demoController) Ping(ctx *gin.Context) {
	data, err := c.demoService.Ping(ctx, dto.PingInput{Name: ctx.Query("name")})
	if err != nil {
		c.ErrorResponse(ctx, err)
		return
	}
	c.SuccessResponse(ctx, data)
}
