package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func registerAliveRoute(rg *gin.RouterGroup) {
	aliveHandler := handlers.NewAliveHandler()

	rg.GET("/alive", aliveHandler.Register)
}
