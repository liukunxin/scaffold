package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	democontroller "go-infra-starter/internal/app/demo/controller"
	demoservice "go-infra-starter/internal/app/demo/service"
	// SCENE_WS_START
	realtime "go-infra-starter/internal/app/realtime"
	// SCENE_WS_END
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
	demoController := democontroller.NewDemoController(
		demoservice.NewDemoService(),
	)
	// SCENE_WS_START
	realtimeHandler := realtime.NewHandler()
	// SCENE_WS_END

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.GetUser)
	api.GET("/demo/ping", demoController.Ping)

	// SCENE_WS_START
	router.GET("/ws", realtimeHandler.Handle)
	// SCENE_WS_END
}
