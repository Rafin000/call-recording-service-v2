package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

// portaoneClient := portaone.NewPortaOneClient()

func registerXDRRoutes(rg *gin.RouterGroup, xdrRepo domain.XDRRepository, portaoneClient portaone.PortaOneClient) {
	xdrHandler := handlers.NewXDRHandler(xdrRepo, portaoneClient)

	// Routes that require normal user authentication
	xdrGroup := rg.Group("/")
	// xdrGroup.Use(middlewares.TokenRequired().
	{
		xdrGroup.POST("/today", xdrHandler.GetXDR)
		xdrGroup.POST("/recording/:i_xdr", xdrHandler.GetCallRecording)
		xdrGroup.POST("/historical", xdrHandler.GetXDRDumps)
		xdrGroup.POST("/historical/:i_xdr", xdrHandler.GetXDRByI_XDR)
	}
}
