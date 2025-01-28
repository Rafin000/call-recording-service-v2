package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AliveHandler struct{}

// NewAliveHandler creates a new instance of AliveHandler.
func NewAliveHandler() *AliveHandler {
	return &AliveHandler{}
}

// Register handles the /alive endpoint and returns a simple health check response.
func (h *AliveHandler) Register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Service is alive!",
	})
}
