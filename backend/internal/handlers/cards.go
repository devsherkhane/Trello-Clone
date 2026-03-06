package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/devsherkhane/trello-clone/internal/database"
	"github.com/devsherkhane/trello-clone/internal/models"
	"github.com/devsherkhane/trello-clone/internal/notifications"
	"github.com/devsherkhane/trello-clone/internal/utils"
	"github.com/gin-gonic/gin"
)

// CreateCard godoc
// @Summary Create a new card
// @Tags Cards
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param card body models.Card true "Card Object"
// @Success 201 {object} models.Card
// @Router /cards [post]
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
        SELECT b.user_id 
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

// UpdateCard modifies a card's title, description, due_date, or label_color
func UpdateCard(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	cardID := c.Param("id")

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
		LabelColor  string `json:"label_color"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Security Check: Ensure the card belongs to a board owned by this user
	var ownerID int
	err := database.DB.QueryRow(`
        SELECT b.user_id 
        FROM boards b 
        JOIN lists l ON b.id = l.board_id 
        JOIN cards c ON l.id = c.list_id 
        WHERE c.id = ?`, cardID).Scan(&ownerID)

	if err != nil || ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this card"})
		return
	}

	// Dynamic update query construction
	query := "UPDATE cards SET title = ?, description = ?, due_date = ?, label_color = ? WHERE id = ?"

	// Handle empty strings for due_date (if frontend sends empty string instead of null)
	var dueDate interface{} = input.DueDate
	if input.DueDate == "" {
		dueDate = nil
	}

	_, err = database.DB.Exec(query, input.Title, input.Description, dueDate, input.LabelColor, cardID)
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
        SELECT b.user_id 
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

	// 1. Fetch the Board ID and verify ownership in one query
	var boardID int
	var ownerID int
	err := database.DB.QueryRow(`
        SELECT l.board_id, b.user_id 
        FROM lists l 
        JOIN boards b ON l.board_id = b.id 
        WHERE l.id = ?`, input.NewListID).Scan(&boardID, &ownerID)

	if err != nil || ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to destination list"})
		return
	}

	// 2. Update the card's list and position
	query := "UPDATE cards SET list_id = ?, position = ? WHERE id = ?"
	_, err = database.DB.Exec(query, input.NewListID, input.NewPosition, cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move card"})
		return
	}

	// 3. Now boardID is defined and can be used for logging
	utils.LogActivity(userID, boardID, fmt.Sprintf("Moved card %s to list %d at position %d", cardID, input.NewListID, input.NewPosition))

	c.JSON(http.StatusOK, gin.H{"message": "Card moved successfully"})
	notifications.GlobalHub.Broadcast(gin.H{"action": "card_moved", "card_id": cardID})
}

// SearchCards looks for cards matching a query with pagination support
func SearchCards(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	queryParam := c.Query("q")

	// Default pagination values
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var cards []models.Card
	// SQL query adding LIMIT and OFFSET for pagination
	sqlQuery := `
		SELECT c.id, c.list_id, c.title, c.description, c.position 
		FROM cards c
		JOIN lists l ON c.list_id = l.id
		JOIN boards b ON l.board_id = b.id
		WHERE b.user_id = ? AND (c.title LIKE ? OR c.description LIKE ?)
		LIMIT ? OFFSET ?`

	searchTerm := "%" + queryParam + "%"
	rows, err := database.DB.Query(sqlQuery, userID, searchTerm, searchTerm, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
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

	// Optional: Return total count for frontend progress bars/pagination UI
	var total int
	countQuery := `
		SELECT COUNT(*) FROM cards c
		JOIN lists l ON c.list_id = l.id
		JOIN boards b ON l.board_id = b.id
		WHERE b.user_id = ? AND (c.title LIKE ? OR c.description LIKE ?)`
	database.DB.QueryRow(countQuery, userID, searchTerm, searchTerm).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"data":  cards,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func AdvancedSearch(c *gin.Context) {
	// 1. Get UserID from the JWT context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 2. Define the Base Query
	query := `
		SELECT DISTINCT c.id, c.list_id, c.title, c.description, c.position, COALESCE(c.due_date, '')
		FROM cards c
		JOIN lists l ON c.list_id = l.id
		JOIN boards b ON l.board_id = b.id
		LEFT JOIN card_labels cl ON c.id = cl.card_id
		WHERE (b.user_id = ? OR EXISTS(
			SELECT 1 FROM board_collaborators 
			WHERE board_id = b.id AND user_id = ?
		))`

	var args []interface{}
	args = append(args, userID.(int), userID.(int))

	// 3. Apply Dynamic Filters

	// Filter by specific Board (Crucial for "this board" search)
	if boardID := c.Query("board_id"); boardID != "" {
		query += " AND b.id = ?"
		args = append(args, boardID)
	}

	// Filter by Label
	if labelID := c.Query("label_id"); labelID != "" {
		query += " AND cl.label_id = ?"
		args = append(args, labelID)
	}

	// Filter by Overdue status
	if isOverdue := c.Query("overdue"); isOverdue == "true" {
		query += " AND c.due_date < NOW() AND c.due_date IS NOT NULL"
	}

	// Filter by Search Term (Title or Description)
	if searchTerm := c.Query("q"); searchTerm != "" {
		query += " AND (c.title LIKE ? OR c.description LIKE ?)"
		likePattern := "%" + searchTerm + "%"
		args = append(args, likePattern, likePattern)
	}

	// 4. Execute the Query
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search"})
		return
	}
	defer rows.Close()

	// 5. Scan results into a slice
	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID,
			&card.ListID,
			&card.Title,
			&card.Description,
			&card.Position,
			&card.DueDate,
		)
		if err != nil {
			log.Printf("Error scanning card: %v", err)
			continue
		}
		cards = append(cards, card)
	}

	if cards == nil {
		cards = []models.Card{}
	}
	c.JSON(http.StatusOK, cards)
}

// BatchUpdateCards handles moving or archiving multiple cards at once
func BatchUpdateCards(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	var input struct {
		CardIDs []int  `json:"card_ids" binding:"required"`
		Action  string `json:"action" binding:"required"` // "move", "archive", "delete"
		ListID  int    `json:"list_id"`                   // Required for "move"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Security: Ensure the user has permission for ALL these cards
	// We use an IN clause and count to verify ownership/collaboration
	var count int
	verifyQuery := `
    SELECT COUNT(DISTINCT c.id) 
    FROM cards c
    JOIN lists l ON c.list_id = l.id
    JOIN boards b ON l.board_id = b.id
    WHERE c.id IN (?) AND (b.user_id = ? OR EXISTS(
        SELECT 1 FROM board_collaborators WHERE board_id = b.id AND user_id = ?
    ))`

	// Use our helper to expand the IN clause
	query, args := PrepareIn(verifyQuery, input.CardIDs, userID, userID)

	err := database.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification failed"})
		return
	}

	if count != len(input.CardIDs) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to modify one or more selected cards"})
		return
	}

	// 2. Perform the Action using the same helper
	if input.Action == "move" {
		updateQuery, updateArgs := PrepareIn("UPDATE cards SET list_id = ? WHERE id IN (?)", input.CardIDs)
		// We need to re-order args because list_id comes first in this query
		finalArgs := append([]interface{}{input.ListID}, updateArgs...)
		_, err = database.DB.Exec(updateQuery, finalArgs...)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch action completed successfully"})
}
func PrepareIn(query string, slice []int, otherArgs ...interface{}) (string, []interface{}) {
	// Create a string of question marks: ?,?,?
	placeholders := make([]string, len(slice))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	// Replace the first (?) found in the query with the new placeholders
	query = strings.Replace(query, "(?)", "("+strings.Join(placeholders, ", ")+")", 1)

	// Combine the slice elements and other arguments into one interface slice
	var args []interface{}
	for _, id := range slice {
		args = append(args, id)
	}
	args = append(args, otherArgs...)

	return query, args
}
