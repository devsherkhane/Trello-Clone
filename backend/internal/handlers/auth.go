package handlers

import (
	"net/http"

	"github.com/devsherkhane/drift/internal/middleware"
	"github.com/devsherkhane/drift/internal/utils"
	"github.com/gin-gonic/gin"
)

func (h *APIHandler) Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email,max=255"`
		Password string `json:"password" binding:"required,min=6,max=128"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": middleware.FormatValidationErrors(err)})
		return
	}

	user, err := h.AuthService.Register(input.Username, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user":    user,
	})
}

func (h *APIHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email,max=255"`
		Password string `json:"password" binding:"required,min=6,max=128"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": middleware.FormatValidationErrors(err)})
		return
	}

	token, user, err := h.AuthService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

func (h *APIHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	user, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *APIHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	var input struct {
		Username string `json:"username" binding:"omitempty,min=3,max=50"`
		Email    string `json:"email" binding:"omitempty,email,max=255"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": middleware.FormatValidationErrors(err)})
		return
	}

	user, err := h.AuthService.UpdateProfile(userID, input.Username, input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

func (h *APIHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email,max=255"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": middleware.FormatValidationErrors(err)})
		return
	}

	err := h.AuthService.ForgotPassword(input.Email)
	if err != nil {
		// Log error, but always return success to prevent enum
		c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
}

func (h *APIHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required,max=255"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": middleware.FormatValidationErrors(err)})
		return
	}

	err := h.AuthService.ResetPassword(input.Token, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful. You can now log in."})
}
