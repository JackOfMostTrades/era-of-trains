package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/wneessen/go-mail"
	"log/slog"
	"net/http"
)

func (server *GameServer) notifyPlayer(gameId string, userId string) error {
	stmt, err := server.db.Prepare("SELECT nickname,discord_user_id,email,email_notifications_enabled,webhooks FROM users WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(userId)
	var nickname string
	var discordUserId sql.NullString
	var email sql.NullString
	var emailNotificationsEnabled int
	var webhooksStr sql.NullString
	err = row.Scan(&nickname, &discordUserId, &email, &emailNotificationsEnabled, &webhooksStr)
	if err != nil {
		return fmt.Errorf("failed to get email for user: %v", err)
	}
	var webhooks []string
	if webhooksStr.Valid {
		err = json.Unmarshal([]byte(webhooksStr.String), &webhooks)
		if err != nil {
			return fmt.Errorf("failed to unmarshal webhooks: %v", err)
		}
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

	if finishedFlag != 0 {
		if email.Valid {
			err = server.sendGameFinishedEmail(gameId, gameName, email.String)
			if err != nil {
				return err
			}
		}
		err = server.sendGameFinishedWebhooks(gameId, gameName, webhooks)
		if err != nil {
			return err
		}
	} else {
		if emailNotificationsEnabled != 0 && email.Valid {
			err = server.sendGameTurnEmail(gameId, gameName, email.String)
			if err != nil {
				return err
			}
		}
		err = server.sendGameTurnWebhooks(discordUserId.String, nickname, gameId, gameName, webhooks)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server *GameServer) sendGameFinishedEmail(gameId string, gameName string, email string) error {
	if server.config.Email == nil {
		slog.Warn("email is not configured; notifications are implicitly disabled")
		return nil
	}
	if server.config.Email.Disabled {
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
	message.Subject("Game finished: (" + gameName + ")")
	message.SetBodyString(mail.TypeTextPlain, "Your game has finished. Follow the link below to see the end state.\n\n"+gameLink)

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

func (server *GameServer) sendGameFinishedWebhooks(gameId string, gameName string, webhooks []string) error {
	gameLink := "https://eot.coderealms.io/games/" + gameId
	message := "Your game [" + gameName + "](" + gameLink + ") has finished."

	for _, webhook := range webhooks {
		body, err := json.Marshal(map[string]interface{}{
			"content": message,
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.Body != nil {
			res.Body.Close()
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("got non-2xx status code from webhook: %d", res.StatusCode)
		}
	}
	return nil
}

func (server *GameServer) sendGameTurnEmail(gameId string, gameName string, email string) error {
	if server.config.Email == nil {
		slog.Warn("email is not configured; notifications are implicitly disabled")
		return nil
	}
	if server.config.Email.Disabled {
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
	message.Subject("It's your turn! (" + gameName + ")")
	message.SetBodyString(mail.TypeTextPlain, "It's your turn! Follow the link below to see the game.\n\n"+gameLink)

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

func (server *GameServer) sendGameTurnWebhooks(discordUserId string, nickname string, gameId string, gameName string, webhooks []string) error {
	gameLink := "https://eot.coderealms.io/games/" + gameId
	var mention string
	if discordUserId != "" {
		mention = "<@" + discordUserId + ">"
	} else {
		mention = nickname
	}
	message := fmt.Sprintf("It is %s's turn at [%s](%s).", mention, gameName, gameLink)

	for _, webhook := range webhooks {
		body, err := json.Marshal(map[string]interface{}{
			"content": message,
			"allowed_mentions": map[string]interface{}{
				"parse": []string{"users"},
			},
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.Body != nil {
			res.Body.Close()
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("got non-2xx status code from webhook: %d", res.StatusCode)
		}
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
		var email sql.NullString
		err = rows.Scan(&id, &email)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		if email.Valid {
			userIdToEmail[id] = email.String
		}
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
