package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	"go-infra-starter/internal/app/user/controller"
	"go-infra-starter/internal/app/user/dao"
	"go-infra-starter/internal/app/user/logic"
	"go-infra-starter/internal/app/user/service"
)

func Setup(router *gin.Engine) {
	router.Use(gin.Recovery(), middlewares.GinTraceMiddleware(), middlewares.HttpLogRecord())

	userController := controller.NewUserController(
		logic.NewUserLogic(
			service.NewUserService(
				dao.NewUserRepo(),
			),
		),
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.GetUser)
}

