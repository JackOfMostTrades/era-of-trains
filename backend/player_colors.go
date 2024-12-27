package main

import "math/rand"

func assignPlayerColors(playerIds map[string]bool) (map[string]int, error) {
	// Randomize player colors
	shuffledColors := make([]int, 6) // 6 colors available to players, regardless of player count
	for idx := 0; idx < len(shuffledColors); idx++ {
		shuffledColors[idx] = idx
	}
	rand.Shuffle(len(shuffledColors), func(i, j int) {
		shuffledColors[i], shuffledColors[j] = shuffledColors[j], shuffledColors[i]
	})
	playerColor := make(map[string]int)
	idx := 0
	for playerId := range playerIds {
		playerColor[playerId] = shuffledColors[idx]
		idx += 1
	}
	return playerColor, nil
}
