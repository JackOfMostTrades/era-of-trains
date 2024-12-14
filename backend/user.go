package main

import (
	"database/sql"
	"fmt"
)

func (server *GameServer) getJoinedUsers(gameId string) (map[string]bool, error) {
	joinedUsers := make(map[string]bool)
	stmt, err := server.db.Prepare("SELECT (player_user_id) FROM game_player_map WHERE game_id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	for rows.Next() {
		var userId string
		err = rows.Scan(&userId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		joinedUsers[userId] = true
	}
	return joinedUsers, nil
}

func (server *GameServer) getUserById(userId string) (*User, error) {
	stmt, err := server.db.Prepare("SELECT nickname,email FROM users WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to excute statement: %v", err)
	}

	var nickname string
	var email string
	err = row.Scan(&nickname, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to excute statement: %v", err)
	}

	return &User{
		Nickname: nickname,
		Email:    email,
		Id:       userId,
	}, nil
}
