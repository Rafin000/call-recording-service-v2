package routes

import (
	"github.com/gin-gonic/gin"
)

func registerAliveRoute(rg *gin.RouterGroup) {
	aliveHandler := handlers.NewAliveHandler()

	rg.GET("/alive", aliveHandler.Register)
}
