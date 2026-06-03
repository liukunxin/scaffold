package controller

import (
	"github.com/gin-gonic/gin"
	kcontroller "github.com/liukunxin/go-infra/pkg/biz/controller"
	"single-starter/internal/app/user/convert"
	"single-starter/internal/app/user/logic"
	"single-starter/internal/app/user/ro"
)

type UserController interface {
	// CreateUser 创建用户并返回用户详情。
	CreateUser(ctx *gin.Context)
	// GetUser 根据用户ID查询用户详情。
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
