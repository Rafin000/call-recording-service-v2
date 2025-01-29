package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func registerXDRRoutes(rg *gin.RouterGroup, xdrRepo domain.XDRRepository) {
	xdrHandler := handlers.NewXDRHandler(xdrRepo)

	// Routes that require normal user authentication
	xdrGroup := rg.Group("/")
	// xdrGroup.Use(middlewares.TokenRequired())
	{
		xdrGroup.POST("/today", xdrHandler.GetXDR)
		xdrGroup.POST("/recording/:i_xdr", xdrHandler.GetCallRecording)
		xdrGroup.POST("/historical", xdrHandler.GetXDRDumps)
		xdrGroup.POST("/historical/:i_xdr", xdrHandler.GetXDRByI_XDR)
	}
}
