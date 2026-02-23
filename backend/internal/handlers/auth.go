package handlers

import (
	"log" // Ensure log is imported
	"net/http"

	"github.com/devsherkhane/trello-clone/internal/auth"
	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Create a new struct specifically for Login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	query := "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)"
	result, err := database.DB.Exec(query, req.Username, req.Email, string(hashed))
	if err != nil {
		log.Printf("Database Insert Error: %v", err) // Log the exact error

		// MySQL error code for duplicate entry is usually handled here
		// For a simpler implementation, we check the general error
		c.JSON(http.StatusConflict, gin.H{
			"error":   "User already exists or registration failed",
			"details": err.Error(),
		})
		return
	}

	// Optional: Get the new user ID
	id, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": id,
	})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	var id int
	var hashedPwd string
	// Ensure the column names (id, password_hash) match your MySQL table exactly
	err := database.DB.QueryRow("SELECT user_id, password_hash FROM users WHERE email = ?", req.Email).Scan(&id, &hashedPwd)
	if err != nil {
		log.Printf("Login Error: %v", err) // CHECK YOUR TERMINAL FOR THIS LOG
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token, err := auth.GenerateToken(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
