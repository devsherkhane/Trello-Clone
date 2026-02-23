package main

import (
	"log"

	"github.com/devsherkhane/trello-clone/internal/auth"
	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

	// 2. Initialize the database connection
	database.InitDB()

	r := gin.Default()

	// 3. CORS Middleware to allow requests from your frontend
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public Routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected Routes (Require JWT)
	api := r.Group("/api")
	api.Use(auth.AuthMiddleware()) 
	{
		// Board Routes
		api.POST("/boards", handlers.CreateBoard)
		api.GET("/boards", handlers.GetBoards)

		// List Routes
		api.POST("/lists", handlers.CreateList)
		api.GET("/boards/:id/lists", handlers.GetListsByBoard)

		// Ping for testing
		api.GET("/ping", func(c *gin.Context) {
			userID := c.MustGet("userID")
			c.JSON(200, gin.H{"user_id": userID})
		})
	}

	// 4. Run the server on port 8080
	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}