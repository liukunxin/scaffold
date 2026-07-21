package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	democontroller "single-starter/internal/app/demo/controller"
	demoservice "single-starter/internal/app/demo/service"
	llmcontroller "single-starter/internal/app/llm/controller"
	llmservice "single-starter/internal/app/llm/service"
	// SCENE_WS_START
	realtime "single-starter/internal/app/realtime"
	// SCENE_WS_END
	"single-starter/internal/app/user/controller"
	"single-starter/internal/app/user/dao"
	"single-starter/internal/app/user/logic"
	"single-starter/internal/app/user/service"
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
	llmController := llmcontroller.NewLLMController(
		llmservice.NewLLMService(),
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.GetUser)
	api.GET("/demo/ping", demoController.Ping)
	// SCENE_WS_START
	router.GET("/ws", realtimeHandler.Handle)
	// SCENE_WS_END
	api.POST("/llm/ping", llmController.Ping)
}
