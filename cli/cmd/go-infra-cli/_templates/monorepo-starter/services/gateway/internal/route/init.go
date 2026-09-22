package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	democontroller "monorepo-starter/services/gateway/internal/app/demo/controller"
	demoservice "monorepo-starter/services/gateway/internal/app/demo/service"
	// SCENE_WS_START
	realtime "monorepo-starter/services/gateway/internal/app/realtime"
	// SCENE_WS_END
)

// Setup 注册 gateway 的全部协议入口。只做传输层粘合，不写业务分支。
func Setup(router *gin.Engine) {
	router.Use(
		gin.Recovery(),
		// 跨域：默认放开所有来源，便于本地前后端联调。
		// 生产请显式白名单：middlewares.CorsMiddleware(middlewares.WithAllowOrigins("https://app.example.com"))
		middlewares.CorsMiddleware(),
		middlewares.GinTraceMiddleware(),
		middlewares.HttpLogRecord(),
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
	api.GET("/demo/ping", demoController.Ping)

	// SCENE_WS_START
	router.GET("/ws", realtimeHandler.Handle)
	// SCENE_WS_END

	// FEATURE_ROUTES_START
	// FEATURE_ROUTES_END
}
