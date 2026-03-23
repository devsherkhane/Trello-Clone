package workers

import (
	"fmt"
	"time"

	"github.com/devsherkhane/drift/internal/database"
	"github.com/devsherkhane/drift/internal/utils"
)

func StartReminderWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			// Find cards due in the next 24 hours that haven't been notified
			rows, _ := database.DB.Query(`
                SELECT c.title, u.email, b.title 
                FROM cards c
                JOIN lists l ON c.list_id = l.id
                JOIN boards b ON l.board_id = b.id
                JOIN users u ON b.owner_id = u.user_id
                WHERE c.due_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 24 HOUR)`)

			for rows.Next() {
				var cardTitle, email, boardTitle string
				rows.Scan(&cardTitle, &email, &boardTitle)
				// Inside your worker logic
				utils.SendEmail(email, "Task Deadline Approaching",
					fmt.Sprintf("Your task '%s' on board '%s' is due soon!", cardTitle, boardTitle))
			}
		}
	}()
}
