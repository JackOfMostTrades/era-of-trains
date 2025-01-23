package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCreateGame(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		MinPlayers: 2,
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(res.Games))
	assert.Equal(t, "game-name", res.Games[0].Name)
	assert.Equal(t, false, res.Games[0].Started)
	assert.Equal(t, false, res.Games[0].Finished)
	assert.Equal(t, 2, res.Games[0].MinPlayers)
	assert.Equal(t, 3, res.Games[0].MaxPlayers)
	assert.Equal(t, "rust_belt", res.Games[0].MapName)
	assert.Equal(t, player1, res.Games[0].OwnerUser.Id)
	assert.Equal(t, 1, len(res.Games[0].JoinedUsers))
	assert.Equal(t, player1, res.Games[0].JoinedUsers[0].Id)
}

func TestCreateGameNameRequired(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	_, err := h.createGame(t, player1, &CreateGameRequest{
		MinPlayers: 2,
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	var httpError *HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.code)

	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(res.Games))
}

func TestCreateGameBadPlayerCounts(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	_, err := h.createGame(t, player1, &CreateGameRequest{
		MinPlayers: 3,
		MaxPlayers: 2,
		MapName:    "rust_belt",
	})
	var httpError *HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.code)

	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(res.Games))
}

func TestJoinGame(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	createRes, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		MinPlayers: 2,
		MaxPlayers: 2,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)
	player2 := h.createUser(t)
	h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(res.Games))
	assert.Equal(t, 2, len(res.Games[0].JoinedUsers))
	assert.ElementsMatch(t, []string{player1, player2},
		[]string{res.Games[0].JoinedUsers[0].Id, res.Games[0].JoinedUsers[1].Id})
}

func TestJoinGameInvalidGameId(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player2 := h.createUser(t)
	_, err := h.joinGame(t, player2, &JoinGameRequest{GameId: "bad-game-id"})
	var httpError *HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.code)
}

func TestJoinGameAlreadyStarted(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	createRes, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		MinPlayers: 2,
		MaxPlayers: 2,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)

	_, err = h.joinGame(t, h.createUser(t), &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.joinGame(t, h.createUser(t), &JoinGameRequest{GameId: "bad-game-id"})
	var httpError *HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.code)
}
