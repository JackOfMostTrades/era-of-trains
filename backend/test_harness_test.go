package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

type TestHarness struct {
	macKey     []byte
	gameServer *GameServer
}

func NewTestHarness(t *testing.T) *TestHarness {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	bootstrapSql, err := os.ReadFile("bootstrap.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(bootstrapSql))
	require.NoError(t, err)

	gameMaps, err := maps.LoadMaps()
	require.NoError(t, err)

	macKey := make([]byte, 32)
	_, err = rand.Read(macKey)
	require.NoError(t, err)

	testHarness := &TestHarness{
		macKey: macKey,
		gameServer: &GameServer{
			config: &Config{
				Authentication: &AuthenticationConfig{},
				MacKey:         base64.StdEncoding.EncodeToString(macKey),
				HttpListenPort: -1,
			},
			db:           db,
			gameMaps:     gameMaps,
			randProvider: &common.CryptoRandProvider{},
		},
	}
	err = testHarness.gameServer.runHttpServer()
	require.NoError(t, err)

	return testHarness
}

func (h *TestHarness) Close() error {
	err := h.gameServer.stopHttpServer()
	if err != nil {
		return err
	}
	err = h.gameServer.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func doApiCall[ReqT any, ResT any](h *TestHarness, t *testing.T, asUserId string, path string, req *ReqT) (*ResT, error) {
	body, err := json.Marshal(req)
	require.NoError(t, err)

	client := &http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d%s", h.gameServer.httpListenPort, path), bytes.NewReader(body))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	if asUserId != "" {
		session, err := h.gameServer.createSession(&Session{UserId: asUserId})
		require.NoError(t, err)
		httpReq.AddCookie(&http.Cookie{
			Name:  "eot-session",
			Value: session,
		})
	}

	res, err := client.Do(httpReq)
	require.NoError(t, err)
	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Logf("failed to read error response body: %v", err)
		}
		return nil, &HttpError{description: string(resBody), code: res.StatusCode}
	}
	defer res.Body.Close()
	resBody := new(ResT)
	err = json.NewDecoder(res.Body).Decode(resBody)
	require.NoError(t, err)
	return resBody, nil
}

func (h *TestHarness) createUser(t *testing.T) string {
	nickname, err := uuid.NewUUID()
	require.NoError(t, err)

	stmt, err := h.gameServer.db.Prepare("INSERT INTO users (id,nickname,email_notifications_enabled,discord_turn_alerts_enabled) VALUES(?,?,0,0)")
	require.NoError(t, err)
	defer stmt.Close()

	userId, err := uuid.NewRandom()
	require.NoError(t, err)
	_, err = stmt.Exec(userId.String(), nickname.String())
	require.NoError(t, err)
	return userId.String()
}

// All of the API methods
func (h *TestHarness) whoami(t *testing.T, asUser string, req *WhoAmIRequest) (*WhoAmIResponse, error) {
	return doApiCall[WhoAmIRequest, WhoAmIResponse](h, t, asUser, "/api/whoami", req)
}

func (h *TestHarness) createGame(t *testing.T, asUser string, req *CreateGameRequest) (*CreateGameResponse, error) {
	return doApiCall[CreateGameRequest, CreateGameResponse](h, t, asUser, "/api/createGame", req)
}

func (h *TestHarness) joinGame(t *testing.T, asUser string, req *JoinGameRequest) (*JoinGameResponse, error) {
	return doApiCall[JoinGameRequest, JoinGameResponse](h, t, asUser, "/api/joinGame", req)
}

func (h *TestHarness) leaveGame(t *testing.T, asUser string, req *LeaveGameRequest) (*LeaveGameResponse, error) {
	return doApiCall[LeaveGameRequest, LeaveGameResponse](h, t, asUser, "/api/leaveGame", req)
}

func (h *TestHarness) startGame(t *testing.T, asUser string, req *StartGameRequest) (*StartGameResponse, error) {
	return doApiCall[StartGameRequest, StartGameResponse](h, t, asUser, "/api/startGame", req)
}

func (h *TestHarness) listGames(t *testing.T, asUser string, req *ListGamesRequest) (*ListGamesResponse, error) {
	return doApiCall[ListGamesRequest, ListGamesResponse](h, t, asUser, "/api/listGames", req)
}

func (h *TestHarness) confirmMove(t *testing.T, asUser string, req *ConfirmMoveRequest) (*ConfirmMoveResponse, error) {
	return doApiCall[ConfirmMoveRequest, ConfirmMoveResponse](h, t, asUser, "/api/confirmMove", req)
}

func (h *TestHarness) viewGame(t *testing.T, asUser string, req *ViewGameRequest) (*ViewGameResponse, error) {
	return doApiCall[ViewGameRequest, ViewGameResponse](h, t, asUser, "/api/viewGame", req)
}

func (h *TestHarness) getGameLogs(t *testing.T, asUser string, req *GetGameLogsRequest) (*GetGameLogsResponse, error) {
	return doApiCall[GetGameLogsRequest, GetGameLogsResponse](h, t, asUser, "/api/getGameLogs", req)
}

func (h *TestHarness) getMyGames(t *testing.T, asUser string, req *GetMyGamesRequest) (*GetMyGamesResponse, error) {
	return doApiCall[GetMyGamesRequest, GetMyGamesResponse](h, t, asUser, "/api/getMyGames", req)
}

func (h *TestHarness) getMyProfile(t *testing.T, asUser string, req *GetMyProfileRequest) (*GetMyProfileResponse, error) {
	return doApiCall[GetMyProfileRequest, GetMyProfileResponse](h, t, asUser, "/api/getMyProfile", req)
}

func (h *TestHarness) setMyProfile(t *testing.T, asUser string, req *SetMyProfileRequest) (*SetMyProfileResponse, error) {
	return doApiCall[SetMyProfileRequest, SetMyProfileResponse](h, t, asUser, "/api/setMyProfile", req)
}

func (h *TestHarness) getGameChat(t *testing.T, asUser string, req *GetGameChatRequest) (*GetGameChatResponse, error) {
	return doApiCall[GetGameChatRequest, GetGameChatResponse](h, t, asUser, "/api/getGameChat", req)
}

func (h *TestHarness) sendGameChat(t *testing.T, asUser string, req *SendGameChatRequest) (*SendGameChatResponse, error) {
	return doApiCall[SendGameChatRequest, SendGameChatResponse](h, t, asUser, "/api/sendGameChat", req)
}

func (h *TestHarness) pollGameStatus(t *testing.T, asUser string, req *PollGameStatusRequest) (*PollGameStatusResponse, error) {
	return doApiCall[PollGameStatusRequest, PollGameStatusResponse](h, t, asUser, "/api/pollGameStatus", req)
}
