package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/utils"
	"github.com/gin-gonic/gin"
)

// AdminTokenRequired is a middleware that checks if the user has a valid admin token.
func AdminTokenRequired(config common.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		payload, err := utils.DecodeAuthToken(token, config)
		if err != nil {
			if err.Error() == "expired" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired Token"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			}
			c.Abort()
			return
		}

		// Check if the role is admin
		if payload.Role == "admin" {
			c.Set("name", payload.Name)
			c.Set("email", payload.Email)
			c.Set("role", payload.Role)

			// Check if ICustomer is non-nil and set it
			if payload.ICustomer != nil {
				c.Set("i_customer", *payload.ICustomer)
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TokenRequired is a middleware that checks if the user has a valid token.
func TokenRequired(config common.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Ensure "Bearer" is in the Authorization header
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with Bearer"})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))

		// Decode and validate the token
		payload, err := utils.DecodeAuthToken(token, config)
		if err != nil {
			// Handle token errors
			if err.Error() == "expired" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired Token"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			}
			c.Abort()
			return
		}

		// Set payload data to context
		c.Set("name", payload.Name)
		c.Set("email", payload.Email)
		c.Set("role", payload.Role)

		// Check if ICustomer is non-nil and set it
		if payload.ICustomer != nil {
			slog.Debug("i_customer from payload", "i_customer", *payload.ICustomer)
			cleanedICustomer := strings.Trim(*payload.ICustomer, `\"`)
			slog.Debug("Cleaned i_customer", "i_customer", cleanedICustomer)

			c.Set("i_customer", cleanedICustomer)
		}
		// Proceed with the next middleware/handler
		c.Next()
	}
}
