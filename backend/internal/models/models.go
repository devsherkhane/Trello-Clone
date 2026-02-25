package models

// User represents the person using the app
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // Scrambled password, hidden from JSON
}

// Board is the top-level container (e.g., "Project Alpha")
type Board struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	OwnerID int    `json:"owner_id"`
}

// List represents a column (e.g., "To Do", "Doing")
type List struct {
	ID       int    `json:"id"`
	BoardID  int    `json:"board_id"`
	Title    string `json:"title"`
	Position int    `json:"position"` // Determines the left-to-right order
}

// Card is the individual task inside a list
type Card struct {
	ID          int    `json:"id"`
	ListID      int    `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	DueDate     string `json:"due_date"`
}

// CardAttachment represents a file linked to a card
type CardAttachment struct {
	ID         int    `json:"id"`
	CardID     int    `json:"card_id"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	UploadedAt string `json:"uploaded_at"`
}

// Comment represents a discussion entry on a card
type Comment struct {
	ID        int    `json:"id"`
	CardID    int    `json:"card_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username,omitempty"` // To show who wrote the comment
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

// Notification represents a real-time alert for a user
type Notification struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	FromUser string `json:"from_user"`
	CardID   int    `json:"card_id"`
}
