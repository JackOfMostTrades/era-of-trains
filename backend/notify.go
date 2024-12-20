package main

import (
	"fmt"
	"github.com/wneessen/go-mail"
	"log/slog"
)

func (server *GameServer) notifyPlayer(gameId string, userId string) error {
	if server.config.Email == nil {
		slog.Warn("email is not configured; notifications are implicitly disabled")
		return nil
	}
	if server.config.Email.Disabled {
		return nil
	}

	stmt, err := server.db.Prepare("SELECT email FROM users WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(userId)
	var email string
	err = row.Scan(&email)
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	stmt, err = server.db.Prepare("SELECT finished,name FROM games WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to get status of game: %v", err)
	}
	defer stmt.Close()
	row = stmt.QueryRow(gameId)
	var finishedFlag int
	var gameName string
	err = row.Scan(&finishedFlag, &gameName)
	if err != nil {
		return fmt.Errorf("failed to get status of game: %v", err)
	}

	message := mail.NewMsg()
	if err := message.From("noreply@eot.coderealms.io"); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(email); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}

	gameLink := "https://eot.coderealms.io/games/" + gameId
	if finishedFlag != 0 {
		message.Subject("Game finished: (" + gameName + ")")
		message.SetBodyString(mail.TypeTextPlain, "Your game has finished. Follow the link below to see the end state.\n\n"+gameLink)
	} else {
		message.Subject("It's your turn! (" + gameName + ")")
		message.SetBodyString(mail.TypeTextPlain, "It's your turn! Follow the link below to see the game.\n\n"+gameLink)
	}

	client, err := mail.NewClient(server.config.Email.SmtpServer, mail.WithPort(server.config.Email.SmtpPort), mail.WithSSL(),
		mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(server.config.Email.SmtpUsername), mail.WithPassword(server.config.Email.SmtpPassword))
	if err != nil {
		return fmt.Errorf("failed to create mail client: %v", err)
	}
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}
	return nil
}
