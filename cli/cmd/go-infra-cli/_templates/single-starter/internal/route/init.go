package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	democontroller "single-starter/internal/app/demo/controller"
	demoservice "single-starter/internal/app/demo/service"
	"single-starter/internal/infra/notifier"
	// SCENE_WS_START
	realtime "single-starter/internal/app/realtime"
	// SCENE_WS_END
	"single-starter/internal/app/user/controller"
	"single-starter/internal/app/user/dao"
	"single-starter/internal/app/user/logic"
	"single-starter/internal/app/user/service"
)

func Setup(router *gin.Engine) {
	router.Use(
		gin.Recovery(),
		// 跨域：默认放开所有来源，便于本地前后端联调。
		// 生产请显式白名单：middlewares.CorsMiddleware(middlewares.WithAllowOrigins("https://app.example.com"))
		middlewares.CorsMiddleware(),
		middlewares.GinTraceMiddleware(),
		middlewares.HttpLogRecord(),
	)

	userController := controller.NewUserController(
		logic.NewUserLogic(
			service.NewUserService(
				dao.NewUserRepo(),
			),
			notifier.NewLogNotifier(),
		),
	)
	demoController := democontroller.NewDemoController(
		demoservice.NewDemoService(),
	)
	// SCENE_WS_START
	realtimeHandler := realtime.NewHandler()
	// SCENE_WS_END

	// 探活接口刻意返回裸 JSON（不套 code/msg/data），供 LB / K8s 直接判定。
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

	// FEATURE_ROUTES_START
	// FEATURE_ROUTES_END
}
