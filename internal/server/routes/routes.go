package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(rg *gin.RouterGroup, mongoDB *mongo.Database, config *common.AppConfig) {
	userRepo := domain.NewUserRepository(mongoDB)
	xdrRepo := domain.NewXDRRepository(mongoDB)

	registerAliveRoute(rg)

	userGroup := rg.Group("/auth")
	registerUserRoutes(userGroup, userRepo)

	xdrGroup := rg.Group("/xdrs")
	registerXDRRoutes(xdrGroup, xdrRepo)
}
