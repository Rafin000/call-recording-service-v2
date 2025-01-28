package routes

import (
	"database/sql"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/gin-gonic/gin"
)

func InitRoutes(rg *gin.RouterGroup, db *sql.DB, config *common.AppConfig) {
	registerAliveRoute(rg)

	userGroup := rg.Group("/users")
	registerUserRoutes(rg, userGroup)
}
