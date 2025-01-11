package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateGame(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		NumPlayers: 2,
		MapName:    "rust_belt",
	})
	res := h.listGames(t, player1, &ListGamesRequest{})
	assert.Equal(t, 1, len(res.Games))
	assert.Equal(t, "game-name", res.Games[0].Name)
	assert.Equal(t, false, res.Games[0].Started)
	assert.Equal(t, false, res.Games[0].Finished)
	assert.Equal(t, 2, res.Games[0].NumPlayers)
	assert.Equal(t, "rust_belt", res.Games[0].MapName)
	assert.Equal(t, player1, res.Games[0].OwnerUser.Id)
	assert.Equal(t, 1, len(res.Games[0].JoinedUsers))
	assert.Equal(t, player1, res.Games[0].JoinedUsers[0].Id)
}

func TestJoinGame(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	createRes := h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		NumPlayers: 2,
		MapName:    "rust_belt",
	})
	player2 := h.createUser(t)
	h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	res := h.listGames(t, player1, &ListGamesRequest{})
	assert.Equal(t, 1, len(res.Games))
	assert.Equal(t, 2, len(res.Games[0].JoinedUsers))
	assert.ElementsMatch(t, []string{player1, player2},
		[]string{res.Games[0].JoinedUsers[0].Id, res.Games[0].JoinedUsers[1].Id})
}
