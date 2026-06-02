package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	infraws "github.com/liukunxin/go-infra/pkg/infra/websocket"
)

type Handler struct {
	server *infraws.Server
	hub    *infraws.Hub
}

func NewHandler() *Handler {
	hub := infraws.NewHub()
	server := infraws.NewServer(
		infraws.Config{},
		infraws.Handlers{
			OnConnect: func(c *infraws.Conn) {
				hub.Register(c)
			},
			OnMessage: func(_ *infraws.Conn, m infraws.Message) {
				// Demo: broadcast message to all connected clients.
				hub.Broadcast(m.Type, m.Data)
			},
			OnClose: func(c *infraws.Conn, _ error) {
				hub.Unregister(c)
			},
		},
		nil,
		func(_ *http.Request) bool { return true },
	)
	return &Handler{
		server: server,
		hub:    hub,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	h.server.ServeHTTP(c.Writer, c.Request)
}

