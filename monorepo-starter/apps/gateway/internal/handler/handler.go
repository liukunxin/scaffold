package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"monorepo-starter/apps/gateway/internal/model"
	"monorepo-starter/apps/gateway/internal/service"
)

type RuntimeHandler struct {
	service *service.RuntimeService
}

func NewRuntimeHandler(service *service.RuntimeService) *RuntimeHandler {
	return &RuntimeHandler{service: service}
}

func (h *RuntimeHandler) HandlePing(c *gin.Context) {
	out, err := h.service.Ping(c.Request.Context(), model.PingQuery{
		Name: c.Query("name"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}
