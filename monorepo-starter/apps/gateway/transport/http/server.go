package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	"go-infra-monorepo-starter/apps/gateway/internal/handler"
	wstransport "go-infra-monorepo-starter/apps/gateway/transport/ws"
)

type Server struct {
	router  *gin.Engine
	handler *handler.RuntimeHandler
	ws      *wstransport.Server
}

func New(router *gin.Engine, handler *handler.RuntimeHandler, ws *wstransport.Server) *Server {
	s := &Server{
		router:  router,
		handler: handler,
		ws:      ws,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(gin.Recovery(), middlewares.GinTraceMiddleware(), middlewares.HttpLogRecord())
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	s.router.GET("/api/v1/runtime/ping", s.handler.HandlePing)
	s.router.GET("/ws", s.ws.Handle)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
