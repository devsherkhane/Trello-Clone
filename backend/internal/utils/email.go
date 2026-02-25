package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendInvitationEmail(toEmail string, boardTitle string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	msg := fmt.Sprintf("Subject: You've been added to a board!\n\nYou have been added as a collaborator to the board: %s", boardTitle)

	auth := smtp.PlainAuth("", from, pass, host)
	return smtp.SendMail(host+":"+port, auth, from, []string{toEmail}, []byte(msg))
}