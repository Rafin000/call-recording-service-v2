package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo domain.UserRepository
}

func NewUserHandler(userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user domain.User
	// Bind JSON body to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Check if the user already exists
	existingUser, err := h.userRepo.GetUserByEmail(ctx, user.Email)
	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User already exists."})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	// Prepare the user struct for creation
	newUser := domain.User{
		Name:      user.Name,
		Email:     user.Email,
		Password:  string(hashedPassword),
		Role:      "user",                          // Default role
		CreatedAt: time.Now().Format(time.RFC3339), // Format time as string
		UpdatedAt: time.Now().Format(time.RFC3339), // Format time as string
	}

	// Save the user to the database
	userID, err := h.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user."})
		return
	}

	// Prepare the response data (similar to Python)
	postData := map[string]interface{}{
		"name":       newUser.Name,
		"email":      newUser.Email,
		"role":       newUser.Role,
		"i_customer": newUser.ICustomer,
		"created_at": newUser.CreatedAt,
		"updated_at": newUser.UpdatedAt,
	}

	// Add the user ID to the response
	postData["_id"] = userID

	// Return the response data with the user ID
	c.JSON(http.StatusOK, postData)
}

func (h *UserHandler) Login(c *gin.Context) {
	var loginData domain.Login
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Check if user exists
	user, _ := h.userRepo.GetUserByEmail(ctx, loginData.Email)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	// Check if password is correct
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Incorrect password."})
		return
	}

	payloads := map[string]interface{}{
		"email":      user.Email,
		"role":       user.Role,
		"name":       user.Name,
		"i_customer": user.ICustomer,
	}

	accessToken, err := utils.GenerateAccessToken(payloads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate access tokens."})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(payloads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate access tokens."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"status":        "success",
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	// Assuming the token is valid
	userEmail := c.MustGet("user_email").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Get user by email
	user, _ := h.userRepo.GetUserByEmail(ctx, userEmail)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}
	payloads := map[string]interface{}{
		"email":      user.Email,
		"role":       user.Role,
		"name":       user.Name,
		"i_customer": user.ICustomer,
	}

	accessToken, err := utils.GenerateAccessToken(payloads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate access tokens."})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(payloads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate access tokens."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"status":        "success",
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userId := c.Param("user_id")

	var updateData domain.UpdateUser
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Check if user exists
	user, _ := h.userRepo.GetUserById(ctx, userId)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	// Check if email already exists
	existingUser, _ := h.userRepo.GetUserByEmail(ctx, updateData.Email)
	if existingUser != nil && existingUser.ID != user.Id {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User with this email already exists."})
		return
	}

	// Update user details
	user.Name = updateData.Name
	user.Email = updateData.Email
	user.UpdatedAt = time.Now()
	if updateData.Role != "" {
		user.Role = updateData.Role
	}
	if updateData.ICustomer != "" {
		user.ICustomer = updateData.ICustomer
	}

	if err := h.userRepo.UpdateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user."})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var passwordData ChangePassword
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Get the email from the context
	email := c.MustGet("user_email").(string)

	// Fetch user from database
	user, _ := h.userRepo.GetUserByEmail(ctx, email)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordData.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	// Update password in the database
	user.Password = string(hashedPassword)
	if err := h.userRepo.UpdateUserPassword(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully."})
}

func (h *UserHandler) AdminChangePassword(c *gin.Context) {
	var passwordData AdminPasswordChange
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Get user by email
	user, _ := h.userRepo.GetUserByEmail(ctx, passwordData.Email)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordData.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	// Update password in the database
	user.Password = string(hashedPassword)
	if err := h.userRepo.UpdateUserPassword(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully."})
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	// DefaultQuery returns a string; we need to convert it to int
	currentPageStr := c.DefaultQuery("current_page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	// Convert string to int
	currentPage, err := strconv.Atoi(currentPageStr)
	if err != nil || currentPage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid current_page value."})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page_size value."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Call the repository function
	paginatedUsers, err := h.userRepo.GetAllUsers(ctx, currentPage, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving users."})
		return
	}

	// Respond with the paginated data
	c.JSON(http.StatusOK, gin.H{
		"users":        paginatedUsers.Users,
		"total_count":  paginatedUsers.TotalCount,
		"total_pages":  paginatedUsers.TotalPages,
		"current_page": paginatedUsers.CurrentPage,
	})
}
