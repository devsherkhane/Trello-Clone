package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
	"github.com/gin-gonic/gin"
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

// UpdateBoard changes the title of an existing board
func UpdateBoard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")
	var input struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	// Verify ownership before updating
	query := "UPDATE boards SET title = ? WHERE id = ? AND owner_id = ?"
	result, err := database.DB.Exec(query, input.Title, boardID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board updated"})
}

// DeleteBoard removes a board and all its associated lists/cards
func DeleteBoard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	query := "DELETE FROM boards WHERE id = ? AND owner_id = ?"
	result, err := database.DB.Exec(query, boardID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board deleted"})
}

func GetBoard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	var b models.Board
	query := "SELECT id, title, owner_id FROM boards WHERE id = ? AND owner_id = ?"
	err := database.DB.QueryRow(query, boardID, userID).Scan(&b.ID, &b.Title, &b.OwnerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		return
	}

	c.JSON(http.StatusOK, b)
}

// GetActivityLogs retrieves the history for a specific board
func GetActivityLogs(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	// Security Check: Ensure user owns the board
	var ownerID int
	err := database.DB.QueryRow("SELECT owner_id FROM boards WHERE id = ?", boardID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to logs"})
		return
	}

	rows, err := database.DB.Query("SELECT action_text, created_at FROM activity_logs WHERE board_id = ? ORDER BY created_at DESC LIMIT 50", boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	defer rows.Close()

	var logs []gin.H
	for rows.Next() {
		var action, createdAt string
		rows.Scan(&action, &createdAt)
		logs = append(logs, gin.H{"action": action, "created_at": createdAt})
	}

	c.JSON(http.StatusOK, logs)
}

func ExportBoardCSV(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	// Security check
	var title string
	err := database.DB.QueryRow("SELECT title FROM boards WHERE id = ? AND owner_id = ?", boardID, userID).Scan(&title)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	rows, _ := database.DB.Query(`
		SELECT l.title, c.title, c.description 
		FROM lists l 
		JOIN cards c ON l.id = c.list_id 
		WHERE l.board_id = ? ORDER BY l.position, c.position`, boardID)
	defer rows.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", title))
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"List", "Card Title", "Description"})

	for rows.Next() {
		var listTitle, cardTitle, desc string
		rows.Scan(&listTitle, &cardTitle, &desc)
		writer.Write([]string{listTitle, cardTitle, desc})
	}
	writer.Flush()
}
