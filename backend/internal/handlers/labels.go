package handlers

import (
	"net/http"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/gin-gonic/gin"
)

func AddLabelToCard(c *gin.Context) {
	var input struct {
		CardID  int `json:"card_id" binding:"required"`
		LabelID int `json:"label_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := database.DB.Exec("INSERT INTO card_labels (card_id, label_id) VALUES (?, ?)", input.CardID, input.LabelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add label"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Label added to card"})
}
