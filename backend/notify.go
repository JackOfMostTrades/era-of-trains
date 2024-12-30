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

	stmt, err := server.db.Prepare("SELECT email,email_notifications_enabled FROM users WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(userId)
	var email string
	var emailNotificationsEnabled int
	err = row.Scan(&email, &emailNotificationsEnabled)
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

	if emailNotificationsEnabled == 0 && finishedFlag == 0 {
		return nil
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
	defer client.Close()
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}
	return nil
}

func (server *GameServer) runDailyNotify() error {
	if server.config.Email == nil {
		slog.Warn("email is not configured; notifications are implicitly disabled")
		return nil
	}
	if server.config.Email.Disabled {
		return nil
	}

	stmt, err := server.db.Prepare("SELECT id,email FROM users")
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return fmt.Errorf("failed to query all users: %v", err)
	}
	defer rows.Close()

	userIdToEmail := make(map[string]string)
	for rows.Next() {
		var id string
		var email string
		err = rows.Scan(&id, &email)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		userIdToEmail[id] = email
	}

	stmt, err = server.db.Prepare("SELECT id,active_player_id FROM games WHERE finished=0")
	if err != nil {
		return fmt.Errorf("failed to get games: %v", err)
	}
	defer stmt.Close()
	rows, err = stmt.Query()
	if err != nil {
		return fmt.Errorf("failed to query all games: %v", err)
	}
	defer rows.Close()

	playerIdToGames := make(map[string][]string)
	for rows.Next() {
		var gameId string
		var playerId string
		err = rows.Scan(&gameId, &playerId)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		if games, ok := playerIdToGames[playerId]; ok {
			playerIdToGames[playerId] = append(games, gameId)
		} else {
			playerIdToGames[playerId] = []string{gameId}
		}
	}

	client, err := mail.NewClient(server.config.Email.SmtpServer, mail.WithPort(server.config.Email.SmtpPort), mail.WithSSL(),
		mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(server.config.Email.SmtpUsername), mail.WithPassword(server.config.Email.SmtpPassword))
	if err != nil {
		return fmt.Errorf("failed to create mail client: %v", err)
	}
	defer client.Close()
	for playerId, games := range playerIdToGames {
		email := userIdToEmail[playerId]
		if email == "" {
			slog.Warn("Not emailing player because they don't have a configured email", "playerId", playerId)
			continue
		}

		message := mail.NewMsg()
		if err := message.From("noreply@eot.coderealms.io"); err != nil {
			return fmt.Errorf("failed to set From address: %s", err)
		}
		if err := message.To(email); err != nil {
			return fmt.Errorf("failed to set To address: %s", err)
		}

		message.Subject("Daily pending games summary")
		content := "It's your turn at the following games! Follow the links below to see them:\n\n"
		for _, gameId := range games {
			content += "https://eot.coderealms.io/games/" + gameId + "\n"
		}

		message.SetBodyString(mail.TypeTextPlain, content)

		slog.Info("Sending daily summary email to player", "playerId", playerId, "email", email, "gameCount", len(games))
		if err := client.DialAndSend(message); err != nil {
			return fmt.Errorf("failed to send mail: %v", err)
		}
	}
	return nil
}
