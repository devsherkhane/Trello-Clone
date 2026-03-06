package handlers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
	"github.com/devsherkhane/trello-clone/internal/notifications"
	"github.com/gin-gonic/gin"
)

// CreateComment adds a new message to a card
func CreateComment(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	var input struct {
		CardID int    `json:"card_id" binding:"required"`
		Text   string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Card ID and text are required"})
		return
	}

	// Security Check: Ensure user has access to the board this card belongs to
	var boardID int
	err := database.DB.QueryRow(`
		SELECT l.board_id FROM cards c
		JOIN lists l ON c.list_id = l.id
		WHERE c.id = ?`, input.CardID).Scan(&boardID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	// Use your existing collaborator check logic
	var ownerID int
	checkQuery := `
		SELECT b.user_id FROM boards b
		LEFT JOIN board_collaborators bc ON b.id = bc.board_id
		WHERE b.id = ? AND (b.user_id = ? OR bc.user_id = ?)`

	err = database.DB.QueryRow(checkQuery, boardID, userID, userID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to comment on this card"})
		return
	}

	query := "INSERT INTO card_comments (card_id, user_id, text) VALUES (?, ?, ?)"
	result, err := database.DB.Exec(query, input.CardID, userID, input.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		return
	}

	// 2. Identify Mentions (@username)
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(input.Text, -1)

	if len(matches) > 0 {
		var senderName string
		database.DB.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&senderName)

		for _, match := range matches {
			mentionedUsername := match[1]

			// Check if the mentioned user exists
			var mentionedUserID int
			err := database.DB.QueryRow("SELECT user_id FROM users WHERE username = ?", mentionedUsername).Scan(&mentionedUserID)

			if err == nil {
				notification := models.Notification{
					Type:     "mention",
					Message:  fmt.Sprintf("%s mentioned you in a comment", senderName),
					FromUser: senderName,
					CardID:   input.CardID,
				}

				// Use the targeted send method instead of broadcast
				notifications.GlobalHub.SendToUser(mentionedUserID, notification)
			}
		}
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "text": input.Text})
}

// GetCommentsByCard retrieves the discussion history for a card
func GetCommentsByCard(c *gin.Context) {
	cardID := c.Param("id")

	var comments []models.Comment
	query := `
		SELECT cc.id, cc.text, cc.created_at, u.username, u.avatar_url 
		FROM card_comments cc
		JOIN users u ON cc.user_id = u.user_id
		WHERE cc.card_id = ?`

	rows, err := database.DB.Query(query, cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var com models.Comment
		if err := rows.Scan(&com.ID, &com.CardID, &com.UserID, &com.Text, &com.CreatedAt, &com.Username); err != nil {
			continue
		}
		comments = append(comments, com)
	}

	c.JSON(http.StatusOK, comments)
}
