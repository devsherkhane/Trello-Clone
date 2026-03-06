package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
	"github.com/devsherkhane/trello-clone/internal/utils"
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
	query := "INSERT INTO boards (title, user_id) VALUES (?, ?)"
	result, err := database.DB.Exec(query, input.Title, userID)
	if err != nil {
		// ADD THIS LINE
		fmt.Println("CREATE BOARD DB ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"title":    input.Title,
		"user_id":  userID,
		"is_owner": true,
		"status":   "accepted",
	})
}

func GetBoards(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	uid := userID.(int)
	archived := c.Query("archived") == "true"
	isArchivedInt := 0
	if archived {
		isArchivedInt = 1
	}

	log.Printf("DEBUG: GetBoards for UID=%d, Archived=%d", uid, isArchivedInt)

	query := `
		SELECT b.id, b.title, b.user_id, (b.user_id = ?) as is_owner, COALESCE(bc.status, 'accepted') as status
		FROM boards b
		LEFT JOIN board_collaborators bc ON b.id = bc.board_id AND bc.user_id = ?
		WHERE (b.user_id = ? OR (bc.user_id = ? AND bc.status IS NOT NULL)) AND b.is_archived = ?
	`

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM boards").Scan(&count)
	log.Printf("DEBUG: Total boards in DB: %d", count)

	rows, err := database.DB.Query(query, uid, uid, uid, uid, isArchivedInt)
	if err != nil {
		log.Printf("GET BOARDS DB ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var boards []gin.H
	for rows.Next() {
		var id, ownerID, isOwner int
		var title, status string
		if err := rows.Scan(&id, &title, &ownerID, &isOwner, &status); err != nil {
			log.Printf("SCAN ERROR: %v", err)
			continue
		}
		boards = append(boards, gin.H{
			"id":       id,
			"title":    title,
			"user_id":  ownerID,
			"is_owner": isOwner == 1,
			"status":   status,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("ROWS ERROR: %v", err)
	}

	log.Printf("DEBUG: GetBoards returning %d boards for user %d", len(boards), uid)
	if boards == nil {
		boards = []gin.H{}
	}
	c.JSON(http.StatusOK, boards)
}

func RespondToInvitation(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")
	var input struct {
		Action string `json:"action" binding:"required"` // "accept" or "decline"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action (accept/decline) is required"})
		return
	}

	if input.Action == "accept" {
		_, err := database.DB.Exec("UPDATE board_collaborators SET status = 'accepted' WHERE board_id = ? AND user_id = ?", boardID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted"})
	} else if input.Action == "decline" {
		_, err := database.DB.Exec("DELETE FROM board_collaborators WHERE board_id = ? AND user_id = ?", boardID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decline invitation"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Invitation declined"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
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
	query := "UPDATE boards SET title = ? WHERE id = ? AND user_id = ?"
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

	query := "DELETE FROM boards WHERE id = ? AND user_id = ?"
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
	// Check if user is owner OR a collaborator
	query := `
        SELECT b.id, b.title, b.user_id 
        FROM boards b
        LEFT JOIN board_collaborators bc ON b.id = bc.board_id
        WHERE b.id = ? AND (b.user_id = ? OR bc.user_id = ?)`

	err := database.DB.QueryRow(query, boardID, userID, userID).Scan(&b.ID, &b.Title, &b.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Fetch Lists
	type CardDetail struct {
		models.Card
		Labels []gin.H `json:"labels,omitempty"`
	}

	type ListWithCards struct {
		models.List
		Cards []CardDetail `json:"cards"`
	}

	var lists []ListWithCards
	listRows, err := database.DB.Query("SELECT id, board_id, title, position FROM lists WHERE board_id = ? ORDER BY position ASC", b.ID)
	if err == nil {
		defer listRows.Close()
		for listRows.Next() {
			var l ListWithCards
			listRows.Scan(&l.ID, &l.BoardID, &l.Title, &l.Position)

			// Fetch Cards for this List
			cardRows, err := database.DB.Query("SELECT id, list_id, title, description, position, due_date FROM cards WHERE list_id = ? ORDER BY position ASC", l.ID)
			if err == nil {
				l.Cards = make([]CardDetail, 0)
				for cardRows.Next() {
					var cd CardDetail
					cardRows.Scan(&cd.ID, &cd.ListID, &cd.Title, &cd.Description, &cd.Position, &cd.DueDate)
					l.Cards = append(l.Cards, cd)
				}
				cardRows.Close()
			}
			lists = append(lists, l)
		}
	}

	// Create a custom response extending models.Board
	response := gin.H{
		"id":      b.ID,
		"title":   b.Title,
		"user_id": b.UserID,
		"lists":   lists,
	}

	if lists == nil {
		response["lists"] = []ListWithCards{}
	}

	c.JSON(http.StatusOK, response)
}

// GetActivityLogs retrieves the history for a specific board
func GetActivityLogs(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	// Security Check: Ensure user owns the board
	var ownerID int
	err := database.DB.QueryRow("SELECT user_id FROM boards WHERE id = ?", boardID).Scan(&ownerID)
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
	err := database.DB.QueryRow("SELECT title FROM boards WHERE id = ? AND user_id = ?", boardID, userID).Scan(&title)
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

func AddCollaborator(c *gin.Context) {
	boardID := c.Param("id")

	var input struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email and role are required"})
		return
	}

	// 1. Find the invited user's ID
	var collaboratorID int
	err := database.DB.QueryRow("SELECT user_id FROM users WHERE email = ?", input.Email).Scan(&collaboratorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 2. Prevent owner from inviting themselves
	var ownerID int
	database.DB.QueryRow("SELECT user_id FROM boards WHERE id = ?", boardID).Scan(&ownerID)
	if collaboratorID == ownerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot invite yourself to your own board"})
		return
	}

	// 3. Insert into board_collaborators
	_, err = database.DB.Exec("INSERT INTO board_collaborators (board_id, user_id, role) VALUES (?, ?, ?)", boardID, collaboratorID, input.Role)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a collaborator"})
		return
	}

	// 3. Send Email Notification
	var boardTitle string
	database.DB.QueryRow("SELECT title FROM boards WHERE id = ?", boardID).Scan(&boardTitle)
	go utils.SendInvitationEmail(input.Email, boardTitle)

	c.JSON(http.StatusOK, gin.H{"message": "Collaborator added and notified"})
}

// ArchiveBoard toggles the archived status
func ArchiveBoard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	boardID := c.Param("id")

	_, err := database.DB.Exec("UPDATE boards SET is_archived = NOT is_archived WHERE id = ? AND user_id = ?", boardID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update board status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board status updated"})
}
