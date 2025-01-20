package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
)

func assignPlayerColors(db *sql.DB, playerIds []string) (map[string]int, error) {
	stmt, err := db.Prepare("SELECT color_preferences FROM users WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	colorPreferences := make(map[string][]int)
	for _, playerId := range playerIds {
		row := stmt.QueryRow(playerId)

		var colorPreferencesStr sql.NullString
		err = row.Scan(&colorPreferencesStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		if colorPreferencesStr.Valid {
			var playerColorPreferences []int
			err = json.Unmarshal([]byte(colorPreferencesStr.String), &playerColorPreferences)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal color preferences: %v", err)
			}
			colorPreferences[playerId] = playerColorPreferences
		} else {
			colorPreferences[playerId] = nil
		}
	}

	playerToColor := make(map[string]int)
	colorToPlayer := make(map[int]string)

	// Randomize player order to break ties
	shuffledPlayers := make([]string, 0, len(playerIds))
	for _, playerId := range playerIds {
		shuffledPlayers = append(shuffledPlayers, playerId)
	}
	rand.Shuffle(len(shuffledPlayers), func(i, j int) {
		shuffledPlayers[i], shuffledPlayers[j] = shuffledPlayers[j], shuffledPlayers[i]
	})

	const NUM_PLAYER_COLORS = 6
	shuffledColors := make([]int, NUM_PLAYER_COLORS)
	for idx := 0; idx < len(shuffledColors); idx++ {
		shuffledColors[idx] = idx
	}
	rand.Shuffle(len(shuffledColors), func(i, j int) {
		shuffledColors[i], shuffledColors[j] = shuffledColors[j], shuffledColors[i]
	})

	for i := 0; i < NUM_PLAYER_COLORS; i++ {
		for _, playerId := range shuffledPlayers {
			if _, ok := playerToColor[playerId]; !ok {
				playerColorPreferences := colorPreferences[playerId]
				if len(playerColorPreferences) > i {
					preferredColor := playerColorPreferences[i]
					if _, ok := colorToPlayer[preferredColor]; !ok {
						playerToColor[playerId] = preferredColor
						colorToPlayer[preferredColor] = playerId
					}
				}
			}
		}
	}

	// After assigning all colors based on preference, now just assign any available
	for _, playerId := range shuffledPlayers {
		if _, ok := playerToColor[playerId]; !ok {
			availableColor := -1
			for _, color := range shuffledColors {
				if _, ok := colorToPlayer[color]; !ok {
					availableColor = color
					break
				}
			}
			if availableColor == -1 {
				return nil, fmt.Errorf("failed to find any available color to assign to player")
			}
			playerToColor[playerId] = availableColor
			colorToPlayer[availableColor] = playerId
		}
	}

	return playerToColor, nil
}
