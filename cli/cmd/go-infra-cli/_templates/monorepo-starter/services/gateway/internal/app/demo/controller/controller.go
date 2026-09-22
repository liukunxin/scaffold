package controller

import (
	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"monorepo-starter/services/gateway/internal/app/demo/dto"
	"monorepo-starter/services/gateway/internal/app/demo/service"
)

type DemoController interface {
	// Ping 三层示例：controller -> service（无 logic）。
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
