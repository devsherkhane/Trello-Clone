package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
)

// CreateList adds a new column to a board
func CreateList(c *gin.Context) {
    userID := c.MustGet("userID").(int)
    var input struct {
        BoardID int    `json:"board_id" binding:"required"`
        Title   string `json:"title" binding:"required"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Board ID and Title are required"})
        return
    }

    // Verify ownership
    var ownerID int
    err := database.DB.QueryRow("SELECT owner_id FROM boards WHERE id = ?", input.BoardID).Scan(&ownerID)
    if err != nil || ownerID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this board"})
        return
    }

    // Atomic Insert to avoid race conditions on 'position'
    query := `
        INSERT INTO lists (board_id, title, position) 
        SELECT ?, ?, COALESCE(MAX(position) + 1, 0) 
        FROM lists WHERE board_id = ?`
    
    result, err := database.DB.Exec(query, input.BoardID, input.Title, input.BoardID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create list"})
        return
    }

    id, _ := result.LastInsertId()
    c.JSON(http.StatusCreated, gin.H{"id": id, "title": input.Title})
}

// GetListsByBoard retrieves all columns for a specific board
func GetListsByBoard(c *gin.Context) {
	boardID := c.Param("id") // We get the board ID from the URL

	var lists []models.List
	query := "SELECT id, board_id, title, position FROM lists WHERE board_id = ? ORDER BY position ASC"

	rows, err := database.DB.Query(query, boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lists"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var l models.List
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Title, &l.Position); err != nil {
			continue
		}
		lists = append(lists, l)
	}

	c.JSON(http.StatusOK, lists)
}