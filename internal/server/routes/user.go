package routes

import (
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(rg *gin.RouterGroup, userRepo domain.UserRepository) {
	userHandler := handlers.NewUserHandler(userRepo)

	// Routes that require Admin authentication
	adminGroup := rg.Group("/admin")
	// adminGroup.Use(middlewares.AdminTokenRequired())
	{
		adminGroup.POST("/create_user", userHandler.CreateUser)
		adminGroup.POST("/update_user/:user_id", userHandler.UpdateUser)
		adminGroup.POST("/get_users", userHandler.GetUsers)
		adminGroup.POST("/change_password", userHandler.AdminChangePassword)
	}

	// Routes that require normal user authentication
	authGroup := rg.Group("/")
	// authGroup.Use(middlewares.TokenRequired())
	{
		authGroup.POST("/change_password", userHandler.ChangePassword)
	}

	// Routes without authentication (Public)
	rg.POST("/login", userHandler.Login)
	rg.POST("/refresh_token", userHandler.RefreshToken)
}
