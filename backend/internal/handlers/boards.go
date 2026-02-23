package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
)

func CreateBoard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	var input struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}
	query := "INSERT INTO boards (title, owner_id) VALUES (?, ?)"
	result, err := database.DB.Exec(query, input.Title, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "title": input.Title})
}

func GetBoards(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	var boards []models.Board
	rows, err := database.DB.Query("SELECT id, title, owner_id FROM boards WHERE owner_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var b models.Board
		rows.Scan(&b.ID, &b.Title, &b.OwnerID)
		boards = append(boards, b)
	}
	c.JSON(http.StatusOK, boards)
}