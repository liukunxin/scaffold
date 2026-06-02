package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	infraws "github.com/liukunxin/go-infra/pkg/infra/websocket"
)

// Server provides a runnable WebSocket reference transport.
type Server struct {
	server *infraws.Server
	hub    *infraws.Hub
}

func NewServer() *Server {
	hub := infraws.NewHub()
	server := infraws.NewServer(
		infraws.Config{},
		infraws.Handlers{
			OnConnect: func(c *infraws.Conn) {
				hub.Register(c)
			},
			OnMessage: func(_ *infraws.Conn, m infraws.Message) {
				// Demo: broadcast all incoming messages.
				hub.Broadcast(m.Type, m.Data)
			},
			OnClose: func(c *infraws.Conn, _ error) {
				hub.Unregister(c)
			},
		},
		nil,
		func(_ *http.Request) bool { return true },
	)
	return &Server{
		server: server,
		hub:    hub,
	}
}

func (s *Server) Handle(c *gin.Context) {
	s.server.ServeHTTP(c.Writer, c.Request)
}
