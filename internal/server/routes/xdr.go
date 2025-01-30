package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/Rafin000/call-recording-service-v2/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

// portaoneClient := portaone.NewPortaOneClient()

func registerXDRRoutes(rg *gin.RouterGroup, xdrRepo domain.XDRRepository, portaoneClient portaone.PortaOneClient, config common.AppConfig) {
	xdrHandler := handlers.NewXDRHandler(xdrRepo, portaoneClient)

	// Routes that require normal user authentication
	xdrGroup := rg.Group("/")
	xdrGroup.Use(middlewares.TokenRequired(config))
	{
		xdrGroup.GET("/today", xdrHandler.GetXDR)
		xdrGroup.GET("/recording/:i_xdr", xdrHandler.GetCallRecording)
		xdrGroup.GET("/historical", xdrHandler.GetXDRDumps)
		xdrGroup.GET("/historical/:i_xdr", xdrHandler.GetXDRByI_XDR)
	}
}
