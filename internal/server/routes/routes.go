package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(rg *gin.RouterGroup, mongoDB *mongo.Database, config *common.AppConfig) {
	userRepo := domain.NewUserRepository(mongoDB)

	registerAliveRoute(rg)

	userGroup := rg.Group("/users")
	userGroup.Use(middlewares.TokenRequired())
	registerUserRoutes(userGroup, userRepo)
}
