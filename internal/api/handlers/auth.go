package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"singbox.arrow.web2/internal/api/middleware"
	"singbox.arrow.web2/internal/storage"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=3"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get stored credentials
	storedUsername, err := storage.GetSetting("username")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}

	storedPasswordHash, err := storage.GetSetting("password_hash")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get password"})
		return
	}

	// Verify credentials
	if req.Username != storedUsername {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "login",
		Detail:    "User logged in: " + req.Username,
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, LoginResponse{
		Token:    token,
		Username: req.Username,
	})
}

func Logout(c *gin.Context) {
	username := c.GetString("username")

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "logout",
		Detail:    "User logged out: " + username,
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify old password
	storedPasswordHash, err := storage.GetSetting("password_hash")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update password
	if err := storage.SetSetting("password_hash", string(newHash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "change_password",
		Detail:    "Password changed",
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
