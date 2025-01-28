package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(rg *gin.RouterGroup, userRepo domain.UserRepository) {
	userHandler := handlers.NewUserHandler(userRepo)
	rg.POST("/admin/create_user", userHandler.CreateUser)
	rg.POST("/login", userHandler.Login)
	rg.POST("/refresh_token", userHandler.RefreshToken)
	rg.POST("/admin/update_user/<string:user_id>", userHandler.UpdateUser)
	rg.POST("/admin/get_users", userHandler.GetUsers)
	rg.POST("/change_password", userHandler.ChangePassword)
	rg.POST("/admin/change_password", userHandler.AdminChangePassword)
}
