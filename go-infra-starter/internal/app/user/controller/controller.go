package controller

import (
	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"go-infra-starter/internal/app/user/convert"
	"go-infra-starter/internal/app/user/logic"
	"go-infra-starter/internal/app/user/ro"
)

type UserController interface {
	CreateUser(ctx *gin.Context)
	GetUser(ctx *gin.Context)
}

type userController struct {
	kcontroller.GinBase
	userLogic logic.UserLogic
}

func NewUserController(userLogic logic.UserLogic) UserController {
	return &userController{userLogic: userLogic}
}

func (c *userController) CreateUser(ctx *gin.Context) {
	var req ro.CreateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ErrorResponse(ctx, err)
		return
	}

	data, err := c.userLogic.CreateUser(ctx, convert.CreateReqToDTO(&req))
	if err != nil {
		c.ErrorResponse(ctx, err)
		return
	}
	c.SuccessResponse(ctx, data)
}

func (c *userController) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	data, err := c.userLogic.GetUser(ctx, id)
	if err != nil {
		c.ErrorResponse(ctx, err)
		return
	}
	c.SuccessResponse(ctx, data)
}

