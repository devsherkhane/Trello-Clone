package utils

import (
	"log"
	"github.com/devsherkhane/drift/internal/database"
)

// LogActivity records a user action in the database
func LogActivity(userID int, boardID int, message string) {
	if database.DB == nil {
		log.Printf("Database not initialized, skipping activity log: %s", message)
		return
	}
	query := "INSERT INTO activity_logs (user_id, board_id, action_text) VALUES (?, ?, ?)"
	_, err := database.DB.Exec(query, userID, boardID, message)
	if err != nil {
		log.Printf("Failed to log activity: %v", err)
	}
}