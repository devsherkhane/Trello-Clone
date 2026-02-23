package handlers

import (
	"net/http"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
	"github.com/gin-gonic/gin"
)

func CreateCard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	var input struct {
		ListID      int    `json:"list_id" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List ID and Title are required"})
		return
	}

	// Security Check: Verify the user owns the board this list belongs to
	var ownerID int
	err := database.DB.QueryRow(`
        SELECT b.owner_id 
        FROM boards b 
        JOIN lists l ON b.id = l.board_id 
        WHERE l.id = ?`, input.ListID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this list"})
		return
	}

	// Insert card with automatic position calculation
	query := `
        INSERT INTO cards (list_id, title, description, position) 
        SELECT ?, ?, ?, COALESCE(MAX(position) + 1, 0) 
        FROM cards WHERE list_id = ?`

	result, err := database.DB.Exec(query, input.ListID, input.Title, input.Description, input.ListID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create card"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "title": input.Title})
}

// GetCardsByList retrieves all cards for a specific list
func GetCardsByList(c *gin.Context) {
	listID := c.Param("id") // Retrieve the list ID from the URL parameter

	var cards []models.Card
	// Select card details ordered by their vertical position
	query := "SELECT id, list_id, title, description, position FROM cards WHERE list_id = ? ORDER BY position ASC"

	rows, err := database.DB.Query(query, listID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cards"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.ListID, &card.Title, &card.Description, &card.Position); err != nil {
			continue
		}
		cards = append(cards, card)
	}

	c.JSON(http.StatusOK, cards)
}

// UpdateCard modifies a card's title or description
func UpdateCard(c *gin.Context) {
    userID := c.MustGet("userID").(int)
    cardID := c.Param("id")
    
    var input struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Security Check: Ensure the card belongs to a board owned by this user
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT b.owner_id 
        FROM boards b 
        JOIN lists l ON b.id = l.board_id 
        JOIN cards c ON l.id = c.list_id 
        WHERE c.id = ?`, cardID).Scan(&ownerID)

    if err != nil || ownerID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this card"})
        return
    }

    query := "UPDATE cards SET title = ?, description = ? WHERE id = ?"
    _, err = database.DB.Exec(query, input.Title, input.Description, cardID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Card updated successfully"})
}

// DeleteCard removes a specific card
func DeleteCard(c *gin.Context) {
    userID := c.MustGet("userID").(int)
    cardID := c.Param("id")

    // Security Check
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT b.owner_id 
        FROM boards b 
        JOIN lists l ON b.id = l.board_id 
        JOIN cards c ON l.id = c.list_id 
        WHERE c.id = ?`, cardID).Scan(&ownerID)

    if err != nil || ownerID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this card"})
        return
    }

    _, err = database.DB.Exec("DELETE FROM cards WHERE id = ?", cardID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete card"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Card deleted"})
}

// MoveCard handles updating a card's position or moving it to a different list
func MoveCard(c *gin.Context) {
    userID := c.MustGet("userID").(int)
    cardID := c.Param("id")

    var input struct {
        NewListID   int `json:"new_list_id" binding:"required"`
        NewPosition int `json:"new_position"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Security Check: Ensure the user owns the board where the card is moving
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT b.owner_id 
        FROM boards b 
        JOIN lists l ON b.id = l.board_id 
        WHERE l.id = ?`, input.NewListID).Scan(&ownerID)

    if err != nil || ownerID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to destination list"})
        return
    }

    // Update the card's list and position
    query := "UPDATE cards SET list_id = ?, position = ? WHERE id = ?"
    _, err = database.DB.Exec(query, input.NewListID, input.NewPosition, cardID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move card"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Card moved successfully"})
}