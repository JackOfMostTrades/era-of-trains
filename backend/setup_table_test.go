package main

import (
	"github.com/JackOfMostTrades/eot/backend/api"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

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
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(res.Games))
}

func TestCreateGamePlayerCountsRequired(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)

	// Test missing min players
	_, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Test missing max players
	_, err = h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-name",
		MinPlayers: 2,
		MapName:    "rust_belt",
	})
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Verify no games were created
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
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
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
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestLeaveGame(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	// Verify both players are in the game
	res, err := h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(res.Games[0].JoinedUsers))

	// Player 2 leaves the game
	_, err = h.leaveGame(t, player2, &LeaveGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	// Verify only player1 remains
	res, err = h.listGames(t, player1, &ListGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(res.Games[0].JoinedUsers))
	assert.Equal(t, player1, res.Games[0].JoinedUsers[0].Id)
}

func TestLeaveGameInvalidGameId(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player := h.createUser(t)
	_, err := h.leaveGame(t, player, &LeaveGameRequest{GameId: "bad-game-id"})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestLeaveGameNotJoined(t *testing.T) {
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
	_, err = h.leaveGame(t, player2, &LeaveGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestLeaveGameAlreadyStarted(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.leaveGame(t, player2, &LeaveGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestDeleteGame(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.deleteGame(t, player1, &DeleteGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.viewGame(t, player1, &ViewGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestDeleteGameNotCreator(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.deleteGame(t, player2, &DeleteGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	viewRes, err := h.viewGame(t, player2, &ViewGameRequest{GameId: createRes.Id})
	require.NoError(t, err)
	assert.False(t, viewRes.Started)
}

func TestStartGameNotCreator(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.startGame(t, player2, &StartGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestStartGameNotEnoughPlayers(t *testing.T) {
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

	_, err = h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestStartGameAlreadyStarted(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	_, err = h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestStartGameInvalidId(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	_, err := h.startGame(t, player1, &StartGameRequest{GameId: "invalid-id"})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
}

func TestStartGameSuccess(t *testing.T) {
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
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: createRes.Id})
	require.NoError(t, err)

	startRes, err := h.startGame(t, player1, &StartGameRequest{GameId: createRes.Id})
	require.NoError(t, err)
	assert.NotNil(t, startRes)

	viewRes, err := h.viewGame(t, player1, &ViewGameRequest{GameId: createRes.Id})
	require.NoError(t, err)
	assert.True(t, viewRes.Started)
}

func TestGetMyGames(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	player2 := h.createUser(t)

	// Create game 1 owned by player1
	game1, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "game-one",
		MinPlayers: 2,
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)

	// Create game 2 owned by player2, player1 joins it
	game2, err := h.createGame(t, player2, &CreateGameRequest{
		Name:       "game-two",
		MinPlayers: 2,
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)
	_, err = h.joinGame(t, player1, &JoinGameRequest{GameId: game2.Id})
	require.NoError(t, err)

	// Create game 3 owned by player2, player1 not involved
	game3, err := h.createGame(t, player2, &CreateGameRequest{
		Name:       "game-three",
		MinPlayers: 2,
		MaxPlayers: 3,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)

	// Get player1's games
	res, err := h.getMyGames(t, player1, &GetMyGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(res.Games))
	assert.ElementsMatch(t, []string{game1.Id, game2.Id}, []string{res.Games[0].Id, res.Games[1].Id})

	// Get player2's games
	res, err = h.getMyGames(t, player2, &GetMyGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(res.Games))
	assert.ElementsMatch(t, []string{game2.Id, game3.Id}, []string{res.Games[0].Id, res.Games[1].Id})
}

func TestGetMyGamesEmpty(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)

	res, err := h.getMyGames(t, player1, &GetMyGamesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(res.Games))
}

func TestGetMyProfile(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	// Create user and verify default profile
	player1 := h.createUser(t)
	profile, err := h.getMyProfile(t, player1, &GetMyProfileRequest{})
	require.NoError(t, err)
	assert.Equal(t, player1, profile.Id)
}

func TestSetMyProfile(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)

	// Update profile
	_, err := h.setMyProfile(t, player1, &SetMyProfileRequest{
		EmailNotificationsEnabled: true,
		DiscordTurnAlertsEnabled:  true,
		ColorPreferences:          []int{1, 2, 3},
		Webhooks:                  []string{"https://example.com/webhook1", "https://example.com/webhook2"},
	})
	require.NoError(t, err)

	// Verify changes
	profile, err := h.getMyProfile(t, player1, &GetMyProfileRequest{})
	require.NoError(t, err)
	assert.Equal(t, player1, profile.Id)
	assert.True(t, profile.EmailNotificationsEnabled)
	assert.True(t, profile.DiscordTurnAlertsEnabled)
	assert.Equal(t, []int{1, 2, 3}, profile.ColorPreferences)
	assert.Equal(t, []string{"https://example.com/webhook1", "https://example.com/webhook2"}, profile.Webhooks)
}

func TestGameChat(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	player2 := h.createUser(t)

	// Create a game
	game, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "chat-test",
		MinPlayers: 2,
		MaxPlayers: 4,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)

	// Player2 joins the game
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: game.Id})
	require.NoError(t, err)

	// Send messages from both players
	_, err = h.sendGameChat(t, player1, &SendGameChatRequest{
		GameId:  game.Id,
		Message: "Hello from player1",
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, err = h.sendGameChat(t, player2, &SendGameChatRequest{
		GameId:  game.Id,
		Message: "Hi player1!",
	})
	require.NoError(t, err)

	// Get chat history
	chatResp, err := h.getGameChat(t, player1, &GetGameChatRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)

	// Verify chat messages
	require.Equal(t, 2, len(chatResp.Messages))
	assert.Equal(t, "Hello from player1", chatResp.Messages[0].Message)
	assert.Equal(t, player1, chatResp.Messages[0].UserId)
	assert.Equal(t, "Hi player1!", chatResp.Messages[1].Message)
	assert.Equal(t, player2, chatResp.Messages[1].UserId)
}

func TestGameChatValidation(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	player2 := h.createUser(t)

	// Create a game
	game, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "chat-test",
		MinPlayers: 2,
		MaxPlayers: 4,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)

	// Test sending message to non-existent game
	_, err = h.sendGameChat(t, player1, &SendGameChatRequest{
		GameId:  "non-existent-game",
		Message: "Hello",
	})
	var httpError *api.HttpError
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Test sending empty message
	_, err = h.sendGameChat(t, player1, &SendGameChatRequest{
		GameId:  game.Id,
		Message: "",
	})
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Test sending message as non-participant
	_, err = h.sendGameChat(t, player2, &SendGameChatRequest{
		GameId:  game.Id,
		Message: "Hello from outsider",
	})
	require.ErrorAs(t, err, &httpError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Test getting chat from non-existent game
	chatResp, err := h.getGameChat(t, player1, &GetGameChatRequest{
		GameId: "non-existent-game",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, len(chatResp.Messages))

	// Test getting chat as non-participant
	_, err = h.getGameChat(t, player2, &GetGameChatRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, len(chatResp.Messages))
}

func TestPollGameStatusWithMoves(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	player2 := h.createUser(t)

	// Create and start a game
	game, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "poll-test",
		MinPlayers: 2,
		MaxPlayers: 4,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: game.Id})
	require.NoError(t, err)
	_, err = h.startGame(t, player1, &StartGameRequest{GameId: game.Id})
	require.NoError(t, err)

	// Get initial status and version
	status1, err := h.pollGameStatus(t, player1, &PollGameStatusRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)
	initialVersion := status1.LastMove

	// Make a move
	time.Sleep(1 * time.Second)
	gameResp, err := h.viewGame(t, player1, &ViewGameRequest{GameId: game.Id})
	require.NoError(t, err)
	_, err = h.confirmMove(t, gameResp.ActivePlayer, &api.ConfirmMoveRequest{
		GameId:       game.Id,
		ActionName:   api.SharesActionName,
		SharesAction: &api.SharesAction{Amount: 0},
	})
	require.NoError(t, err)

	// Poll should show new version and updated game state
	status2, err := h.pollGameStatus(t, player1, &PollGameStatusRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)
	assert.NotEqual(t, initialVersion, status2.LastMove)
}

func TestPollGameStatusWithChat(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	player2 := h.createUser(t)

	// Create and start a game
	game, err := h.createGame(t, player1, &CreateGameRequest{
		Name:       "poll-test",
		MinPlayers: 2,
		MaxPlayers: 4,
		MapName:    "rust_belt",
	})
	require.NoError(t, err)
	_, err = h.joinGame(t, player2, &JoinGameRequest{GameId: game.Id})
	require.NoError(t, err)

	// Get initial status and version
	status1, err := h.pollGameStatus(t, player1, &PollGameStatusRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)
	initialVersion := status1.LastChat

	// Send a chat message
	_, err = h.sendGameChat(t, player1, &SendGameChatRequest{
		GameId:  game.Id,
		Message: "Hello!",
	})
	require.NoError(t, err)

	// Poll should show new version and updated chat
	status2, err := h.pollGameStatus(t, player1, &PollGameStatusRequest{
		GameId: game.Id,
	})
	require.NoError(t, err)
	assert.NotEqual(t, initialVersion, status2.LastChat)
	assert.Greater(t, status2.LastChat, status1.LastChat)
}
