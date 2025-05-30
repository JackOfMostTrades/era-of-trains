package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/api"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/google/uuid"
)

type GameServer struct {
	config         *Config
	db             *sql.DB
	gameMaps       map[string]maps.GameMap
	randProvider   common.RandProvider
	httpServer     *http.Server
	httpListenPort int
}

type User struct {
	Nickname string `json:"nickname"`
	Id       string `json:"id"`
}

type LoginRequest struct {
	Provider    string `json:"provider"`
	AccessToken string `json:"accessToken"`
	// For development purpsose only. If devmode is enabled, sign-in with the given nickname, bypassing actual authentication
	DevNickname string `json:"devNickname"`
}

type LoginResponse struct {
	RegistrationRequired bool `json:"registrationRequired"`
}

func (server *GameServer) Close() error {
	err := server.db.Close()
	if err != nil {
		return err
	}
	return nil
}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

type DiscordUserInfo struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	userInfoReq, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	userInfoReq.Header.Add("Authorization", "Bearer "+accessToken)
	userInfoRes, err := http.DefaultClient.Do(userInfoReq)
	if err != nil {
		return nil, err
	}
	if userInfoRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info from access token: %d", userInfoRes.StatusCode)
	}
	defer userInfoRes.Body.Close()

	userInfoResponse := new(GoogleUserInfo)
	err = json.NewDecoder(userInfoRes.Body).Decode(&userInfoResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %v", err)
	}

	return userInfoResponse, nil
}

func getDiscordUserInfo(accessToken string) (*DiscordUserInfo, error) {
	userInfoReq, err := http.NewRequest(http.MethodGet, "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, err
	}
	userInfoReq.Header.Add("Authorization", "Bearer "+accessToken)
	userInfoRes, err := http.DefaultClient.Do(userInfoReq)
	if err != nil {
		return nil, err
	}
	if userInfoRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info from access token: %d", userInfoRes.StatusCode)
	}
	defer userInfoRes.Body.Close()

	userInfoResponse := new(DiscordUserInfo)
	err = json.NewDecoder(userInfoRes.Body).Decode(&userInfoResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %v", err)
	}

	return userInfoResponse, nil
}

func (server *GameServer) login(ctx *RequestContext, req *LoginRequest) (resp *LoginResponse, err error) {
	var userId string
	if server.config.Authentication.EnableDevLogin && req.DevNickname != "" {
		stmt, err := server.db.Prepare("SELECT id FROM users WHERE nickname=?")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %v", err)
		}
		defer stmt.Close()
		row := stmt.QueryRow(req.DevNickname)
		err = row.Scan(&userId)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, &api.HttpError{fmt.Sprintf("invalid nickname: %s", req.DevNickname), http.StatusBadRequest}
			}
			return nil, fmt.Errorf("failed to lookup user: %v", err)
		}
	} else {
		if req.AccessToken == "" {
			return nil, &api.HttpError{"Missing access token", http.StatusBadRequest}
		}
		if req.Provider == "" {
			return nil, &api.HttpError{"Missing provider parameter", http.StatusBadRequest}
		}

		var getUserQuery string
		var getUserQueryParam string
		if req.Provider == "google" {
			userInfoResponse, err := getGoogleUserInfo(req.AccessToken)
			if err != nil {
				return nil, fmt.Errorf("failed to verify access token: %v", err)
			}
			getUserQuery = "SELECT id FROM users WHERE google_user_id=?"
			getUserQueryParam = userInfoResponse.Id
		} else if req.Provider == "discord" {
			userInfoResponse, err := getDiscordUserInfo(req.AccessToken)
			if err != nil {
				return nil, fmt.Errorf("failed to verify access token: %v", err)
			}
			getUserQuery = "SELECT id FROM users WHERE discord_user_id=?"
			getUserQueryParam = userInfoResponse.Id
		} else {
			return nil, &api.HttpError{fmt.Sprintf("Unsupported provider parameter: %s", req.Provider), http.StatusBadRequest}
		}

		stmt, err := server.db.Prepare(getUserQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %v", err)
		}
		defer stmt.Close()
		row := stmt.QueryRow(getUserQueryParam)

		err = row.Scan(&userId)
		if err != nil {
			if err == sql.ErrNoRows {
				return &LoginResponse{RegistrationRequired: true}, nil
			} else {
				return nil, fmt.Errorf("failed to excute statement: %v", err)
			}
		}
	}

	session := &Session{
		UserId: userId,
	}
	sessionStr, err := server.createSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session cookie: %v", err)
	}

	http.SetCookie(ctx.HttpResponse, &http.Cookie{
		Name:     "eot-session",
		Value:    sessionStr,
		Secure:   !server.config.Authentication.DisableSecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	return &LoginResponse{}, nil
}

type RegisterRequest struct {
	Provider    string `json:"provider"`
	AccessToken string `json:"accessToken"`
	Nickname    string `json:"nickname"`
}

type RegisterResponse struct {
}

func (server *GameServer) register(ctx *RequestContext, req *RegisterRequest) (resp *RegisterResponse, err error) {
	if req.Provider == "" {
		return nil, &api.HttpError{"Missing provider parameter", http.StatusBadRequest}
	}
	if req.AccessToken == "" {
		return nil, &api.HttpError{"Missing access token", http.StatusBadRequest}
	}
	r, err := regexp.Compile("^[a-zA-Z0-9]+$")
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %v", err)
	}
	if !r.MatchString(req.Nickname) {
		return nil, &api.HttpError{fmt.Sprintf("invalid nickname: %s", req.Nickname), http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("SELECT id FROM users WHERE nickname=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.Nickname)
	var existingUserId string
	err = row.Scan(&existingUserId)
	if err != sql.ErrNoRows {
		return nil, &api.HttpError{fmt.Sprintf("user with nickname %s already exists", req.Nickname), http.StatusBadRequest}
	}

	var email sql.NullString
	var googleUserId sql.NullString
	var discordUserId sql.NullString
	if req.Provider == "google" {
		userInfoResponse, err := getGoogleUserInfo(req.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to verify access token: %v", err)
		}
		googleUserId.Valid = true
		googleUserId.String = userInfoResponse.Id
		if userInfoResponse.VerifiedEmail && userInfoResponse.Email != "" {
			email.Valid = true
			email.String = userInfoResponse.Email
		}
	} else if req.Provider == "discord" {
		userInfoResponse, err := getDiscordUserInfo(req.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to verify access token: %v", err)
		}
		discordUserId.Valid = true
		discordUserId.String = userInfoResponse.Id
		if userInfoResponse.Verified && userInfoResponse.Email != "" {
			email.Valid = true
			email.String = userInfoResponse.Email
		}
	} else {
		return nil, &api.HttpError{fmt.Sprintf("unsupported provider parameter: %s", req.Provider), http.StatusBadRequest}
	}

	stmt, err = server.db.Prepare("INSERT INTO users (id,nickname,email,email_notifications_enabled,google_user_id,discord_user_id,discord_turn_alerts_enabled) VALUES(?,?,?,1,?,?,0)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	userId, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %v", err)
	}
	_, err = stmt.Exec(userId.String(), req.Nickname, email, googleUserId, discordUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	session := &Session{
		UserId: userId.String(),
	}
	sessionStr, err := server.createSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session cookie: %v", err)
	}

	http.SetCookie(ctx.HttpResponse, &http.Cookie{
		Name:     "eot-session",
		Value:    sessionStr,
		Secure:   !server.config.Authentication.DisableSecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	return &RegisterResponse{}, nil
}

type LinkProfileRequest struct {
	Provider    string `json:"provider"`
	AccessToken string `json:"accessToken"`
}

type LinkProfileResponse struct {
}

func (server *GameServer) linkProfile(ctx *RequestContext, req *LinkProfileRequest) (resp *LinkProfileResponse, err error) {
	if req.AccessToken == "" {
		return nil, &api.HttpError{"Missing access token", http.StatusBadRequest}
	}
	if req.Provider == "" {
		return nil, &api.HttpError{"Missing provider parameter", http.StatusBadRequest}
	}
	if req.Provider == "google" {
		userInfo, err := getGoogleUserInfo(req.AccessToken)
		if err != nil {
			return nil, &api.HttpError{fmt.Sprintf("Failed to verify access token: %v", err), http.StatusBadRequest}
		}
		if userInfo.Id != "" {
			stmt, err := server.db.Prepare("UPDATE users SET google_user_id=? WHERE id=?")
			if err != nil {
				return nil, fmt.Errorf("failed to prepare statement: %v", err)
			}
			defer stmt.Close()
			_, err = stmt.Exec(userInfo.Id, ctx.User.Id)
			if err != nil {
				return nil, fmt.Errorf("failed to execute stament: %v", err)
			}
		}
	} else if req.Provider == "discord" {
		userInfo, err := getDiscordUserInfo(req.AccessToken)
		if err != nil {
			return nil, &api.HttpError{fmt.Sprintf("Failed to verify access token: %v", err), http.StatusBadRequest}
		}
		if userInfo.Id != "" {
			stmt, err := server.db.Prepare("UPDATE users SET discord_user_id=? WHERE id=?")
			if err != nil {
				return nil, fmt.Errorf("failed to prepare statement: %v", err)
			}
			defer stmt.Close()
			_, err = stmt.Exec(userInfo.Id, ctx.User.Id)
			if err != nil {
				return nil, fmt.Errorf("failed to execute stament: %v", err)
			}
		}
	} else {
		return nil, &api.HttpError{fmt.Sprintf("Unsupported provider parameter: %s", req.Provider), http.StatusBadRequest}
	}

	return &LinkProfileResponse{}, nil
}

type LogoutRequest struct{}

type LogoutResponse struct {
}

func (server *GameServer) logout(ctx *RequestContext, req *LogoutRequest) (resp *LogoutResponse, err error) {
	http.SetCookie(ctx.HttpResponse, &http.Cookie{
		Name:     "eot-session",
		Value:    "",
		Secure:   !server.config.Authentication.DisableSecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	return &LogoutResponse{}, nil
}

type CreateGameRequest struct {
	Name       string `json:"name"`
	MinPlayers int    `json:"minPlayers"`
	MaxPlayers int    `json:"maxPlayers"`
	MapName    string `json:"mapName"`
	InviteOnly bool   `json:"inviteOnly"`
}

type CreateGameResponse struct {
	Id string `json:"id"`
}

func (server *GameServer) createGame(ctx *RequestContext, req *CreateGameRequest) (resp *CreateGameResponse, err error) {
	if req.Name == "" {
		return nil, &api.HttpError{"missing name parameter", http.StatusBadRequest}
	}
	if req.MinPlayers == 0 || req.MaxPlayers == 0 || req.MinPlayers > req.MaxPlayers {
		return nil, &api.HttpError{"invalid or missing minPlayers/maxPlayers parameter", http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("INSERT INTO games (id,created_at,name,min_players,max_players,map_name,owner_user_id,started,finished,invite_only) VALUES (?,?,?,?,?,?,?,0,0,?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %v", err)
	}
	_, err = stmt.Exec(id.String(), time.Now().Unix(), req.Name, req.MinPlayers, req.MaxPlayers, req.MapName, ctx.User.Id, boolToInt(req.InviteOnly))
	if err != nil {
		return nil, fmt.Errorf("failed to insert game row: %v", err)
	}

	stmt, err = server.db.Prepare("INSERT INTO game_player_map (game_id,player_user_id) VALUES(?,?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(id.String(), ctx.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert game_player_map row: %v", err)
	}

	return &CreateGameResponse{Id: id.String()}, nil
}

type JoinGameRequest struct {
	GameId string `json:"gameId"`
}
type JoinGameResponse struct {
}

func (server *GameServer) joinGame(ctx *RequestContext, req *JoinGameRequest) (resp *JoinGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT min_players,max_players,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var minPlayers int
	var maxPlayers int
	var startedFlag int
	err = row.Scan(&minPlayers, &maxPlayers, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &api.HttpError{"game has already started", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if _, ok := joinedUsers[ctx.User.Id]; ok {
		return nil, &api.HttpError{"you have already joined this game", http.StatusBadRequest}
	}
	if len(joinedUsers) >= maxPlayers {
		return nil, &api.HttpError{"this game is already full", http.StatusBadRequest}
	}

	stmt, err = server.db.Prepare("INSERT INTO game_player_map (game_id,player_user_id) VALUES (?,?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(req.GameId, ctx.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// If enough players have joined, mark the owner as the "active player" to indicate the game is waiting on their action
	if len(joinedUsers) >= minPlayers-1 {
		stmt, err = server.db.Prepare("UPDATE games SET active_player_id=owner_user_id WHERE id=?")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare query: %v", err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return &JoinGameResponse{}, nil
}

type LeaveGameRequest struct {
	GameId string `json:"gameId"`
}
type LeaveGameResponse struct {
}

func (server *GameServer) leaveGame(ctx *RequestContext, req *LeaveGameRequest) (resp *LeaveGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,min_players,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var minPlayers int
	var startedFlag int
	err = row.Scan(&ownerUserId, &minPlayers, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &api.HttpError{"game has already started", http.StatusBadRequest}
	}
	if ownerUserId == ctx.User.Id {
		return nil, &api.HttpError{"you cannot leave a game that you created", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if _, ok := joinedUsers[ctx.User.Id]; !ok {
		return nil, &api.HttpError{"you are not joined to this game", http.StatusBadRequest}
	}

	stmt, err = server.db.Prepare("DELETE FROM game_player_map WHERE game_id=? AND player_user_id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(req.GameId, ctx.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// If joined users have dropped back below the min required, make sure the "active player" for the game is set back
	// to NULL to remove indication that game is waiting on owner action.
	if len(joinedUsers)-1 < minPlayers {
		stmt, err = server.db.Prepare("UPDATE games SET active_player_id=NULL WHERE id=?")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare query: %v", err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return &LeaveGameResponse{}, nil
}

type DeleteGameRequest struct {
	GameId string `json:"gameId"`
}
type DeleteGameResponse struct {
}

func (server *GameServer) deleteGame(ctx *RequestContext, req *LeaveGameRequest) (resp *DeleteGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var startedFlag int
	err = row.Scan(&ownerUserId, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &api.HttpError{"game has already started", http.StatusBadRequest}
	}
	if ownerUserId != ctx.User.Id {
		return nil, &api.HttpError{"you cannot delete a game unless you created it", http.StatusBadRequest}
	}

	trx, err := server.db.BeginTx(ctx.HttpRequest.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin sql transaction: %v", err)
	}
	defer trx.Rollback()
	for _, query := range []string{
		"DELETE FROM game_chat WHERE game_id=?",
		"DELETE FROM game_log WHERE game_id=?",
		"DELETE FROM game_player_map WHERE game_id=?",
		"DELETE FROM games WHERE id=?",
	} {
		stmt, err = trx.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare query %s: %v", query, err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %v", err)
		}
	}
	err = trx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &DeleteGameResponse{}, nil
}

type StartGameRequest struct {
	GameId string `json:"gameId"`
}
type StartGameResponse struct {
}

func (server *GameServer) startGame(ctx *RequestContext, req *StartGameRequest) (resp *StartGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,min_players,max_players,map_name,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var minPlayers int
	var maxPlayers int
	var mapName string
	var startedFlag int
	err = row.Scan(&ownerUserId, &minPlayers, &maxPlayers, &mapName, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &api.HttpError{"game has already started", http.StatusBadRequest}
	}
	if ownerUserId != ctx.User.Id {
		return nil, &api.HttpError{"you are not the owner of this game", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if len(joinedUsers) < minPlayers {
		return nil, &api.HttpError{"game does not have enough joined players yet", http.StatusBadRequest}
	}
	if len(joinedUsers) > maxPlayers {
		return nil, &api.HttpError{"game has too many joined players", http.StatusBadRequest}
	}

	stmt, err = server.db.Prepare("UPDATE games SET started=1 WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Randomize initial player order
	playerOrder := make([]string, 0, len(joinedUsers))
	for userId := range joinedUsers {
		playerOrder = append(playerOrder, userId)
	}
	rand.Shuffle(len(playerOrder), func(i, j int) {
		playerOrder[i], playerOrder[j] = playerOrder[j], playerOrder[i]
	})

	// Randomize player colors
	playerColor, err := assignPlayerColors(server.db, joinedUsers)

	// Setup initial game state
	gameState := &common.GameState{
		PlayerOrder:       playerOrder,
		PlayerColor:       playerColor,
		PlayerShares:      make(map[string]int),
		PlayerLoco:        make(map[string]int),
		PlayerIncome:      make(map[string]int),
		PlayerActions:     make(map[string]common.SpecialAction),
		PlayerCash:        make(map[string]int),
		AuctionState:      make(map[string]int),
		GamePhase:         common.SHARES_GAME_PHASE,
		TurnNumber:        1,
		MovingGoodsRound:  0,
		PlayerHasDoneLoco: make(map[string]bool),
		Links:             nil,
		Urbanizations:     nil,
		CubeBag: map[common.Color]int{
			common.BLACK:  16,
			common.RED:    20,
			common.YELLOW: 20,
			common.BLUE:   20,
			common.PURPLE: 20,
		},
		Cubes:           nil,
		GoodsGrowth:     make([][]common.Color, 20),
		ProductionCubes: nil,
	}
	for _, userId := range playerOrder {
		gameState.PlayerShares[userId] = 2
		gameState.PlayerLoco[userId] = 1
		gameState.PlayerIncome[userId] = 0
		gameState.PlayerCash[userId] = 10
	}

	gameMap := server.gameMaps[mapName]
	if gameMap == nil {
		return nil, fmt.Errorf("failed to lookup map: %s", mapName)
	}

	err = gameMap.PopulateStartingCubes(gameState, server.randProvider)

	// Populate the goods growth table
	for i := 0; i < 12; i++ {
		gameState.GoodsGrowth[i] = make([]common.Color, 3)
		if gameMap.GetCityHexForGoodsGrowth(i).X >= 0 {
			for j := 0; j < 3; j++ {
				cube, err := gameState.DrawCube(server.randProvider)
				if err != nil {
					return nil, fmt.Errorf("failed to draw cube: %v", err)
				}
				gameState.GoodsGrowth[i][j] = cube
			}
		}
	}
	for i := 12; i < 20; i++ {
		gameState.GoodsGrowth[i] = make([]common.Color, 2)
		for j := 0; j < 2; j++ {
			cube, err := gameState.DrawCube(server.randProvider)
			if err != nil {
				return nil, fmt.Errorf("failed to draw cube: %v", err)
			}
			gameState.GoodsGrowth[i][j] = cube
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to populate initial board cubes: %v", err)
	}
	err = gameMap.PostSetupHook(gameState, server.randProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to populate initial board cubes: %v", err)
	}

	gameStateStr, err := json.Marshal(gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game state: %v", err)
	}
	stmt, err = server.db.Prepare("UPDATE games SET active_player_id=?,game_state=? WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(playerOrder[0], string(gameStateStr), req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Log the start of the game
	stmt, err = server.db.Prepare("INSERT INTO game_log (game_id,timestamp,user_id,action,description,new_active_player,new_game_state,reversible) VALUES(?, ?, ?, ?, ?, ?, ?, 0)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(req.GameId, time.Now().Unix(), ctx.User.Id, sql.NullString{},
		"The game has started!", playerOrder[0], string(gameStateStr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Notify first player it is their turn
	err = server.notifyPlayer(req.GameId, playerOrder[0])
	if err != nil {
		return nil, fmt.Errorf("failed to notify user it's their turn: %v", err)
	}

	return &StartGameResponse{}, nil
}

type ViewGameRequest struct {
	GameId string `json:"gameId"`
}
type ViewGameResponse struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Started      bool              `json:"started"`
	Finished     bool              `json:"finished"`
	MinPlayers   int               `json:"minPlayers"`
	MaxPlayers   int               `json:"maxPlayers"`
	MapName      string            `json:"mapName"`
	OwnerUser    *User             `json:"ownerUser"`
	ActivePlayer string            `json:"activePlayer"`
	JoinedUsers  []*User           `json:"joinedUsers"`
	GameState    *common.GameState `json:"gameState"`
	InviteOnly   bool              `json:"inviteOnly"`
}

func (server *GameServer) viewGame(ctx *RequestContext, req *ViewGameRequest) (resp *ViewGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT name,owner_user_id,min_players,max_players,map_name,started,finished,game_state,active_player_id,invite_only FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var name string
	var ownerUserId string
	var minPlayers int
	var maxPlayers int
	var mapName string
	var startedFlag int
	var finishedFlag int
	var gameStateStr sql.NullString
	var activePlayerStr sql.NullString
	var inviteOnlyFlag int
	err = row.Scan(&name, &ownerUserId, &minPlayers, &maxPlayers, &mapName, &startedFlag, &finishedFlag, &gameStateStr, &activePlayerStr, &inviteOnlyFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	owner, err := server.getUserById(ownerUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch owner: %v", err)
	}

	joinedUserIds, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch joined users: %v", err)
	}
	joinedUsers := make([]*User, 0, len(joinedUserIds))
	for userId := range joinedUserIds {
		user, err := server.getUserById(userId)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %s: %v", userId, err)
		}
		joinedUsers = append(joinedUsers, user)
	}

	var gameState *common.GameState
	if gameStateStr.Valid {
		gameState = new(common.GameState)
		err = json.Unmarshal([]byte(gameStateStr.String), gameState)
		if err != nil {
			return nil, fmt.Errorf("failed to parse game state: %v", err)
		}
	}

	res := &ViewGameResponse{
		Id:           req.GameId,
		Name:         name,
		Started:      startedFlag != 0,
		Finished:     finishedFlag != 0,
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
		MapName:      mapName,
		OwnerUser:    owner,
		ActivePlayer: activePlayerStr.String,
		JoinedUsers:  joinedUsers,
		GameState:    gameState,
		InviteOnly:   inviteOnlyFlag != 0,
	}

	return res, nil
}

type GameSummary struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Started      bool    `json:"started"`
	Finished     bool    `json:"finished"`
	MinPlayers   int     `json:"minPlayers"`
	MaxPlayers   int     `json:"maxPlayers"`
	MapName      string  `json:"mapName"`
	ActivePlayer string  `json:"activePlayer"`
	OwnerUser    *User   `json:"ownerUser"`
	JoinedUsers  []*User `json:"joinedUsers"`
}

type ListGamesRequest struct {
}
type ListGamesResponse struct {
	Games []*GameSummary `json:"games"`
}

func (server *GameServer) listGames(ctx *RequestContext, req *ListGamesRequest) (resp *ListGamesResponse, err error) {
	stmt, err := server.db.Prepare("SELECT id,name,owner_user_id,min_players,max_players,map_name,started,finished,active_player_id FROM games WHERE invite_only=0 ORDER by created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()
	games, err := server.getGamesSummaries(rows)
	if err != nil {
		return nil, err
	}

	return &ListGamesResponse{
		Games: games,
	}, nil
}

type GameLogEntry struct {
	Timestamp   int    `json:"timestamp"`
	UserId      string `json:"userId"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Reversible  bool   `json:"reversible"`
}

type GetGameLogsRequest struct {
	GameId string `json:"gameId"`
}
type GetGameLogsResponse struct {
	Logs []*GameLogEntry `json:"logs"`
}

func (server *GameServer) getGameLogs(ctx *RequestContext, req *GetGameLogsRequest) (resp *GetGameLogsResponse, err error) {
	stmt, err := server.db.Prepare("SELECT timestamp,user_id,action,description,reversible FROM game_log WHERE game_id=? ORDER BY timestamp ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var entries []*GameLogEntry
	for rows.Next() {
		var timestamp int
		var userId string
		var action sql.NullString
		var description string
		var reversibleFlag int
		err = rows.Scan(&timestamp, &userId, &action, &description, &reversibleFlag)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		entries = append(entries, &GameLogEntry{
			Timestamp:   timestamp,
			UserId:      userId,
			Action:      action.String,
			Description: description,
			Reversible:  reversibleFlag != 0,
		})
	}

	return &GetGameLogsResponse{
		Logs: entries,
	}, nil
}

func (server *GameServer) getGamesSummaries(rows *sql.Rows) ([]*GameSummary, error) {
	var games []*GameSummary

	for rows.Next() {
		var id string
		var name string
		var ownerUserId string
		var minPlayers int
		var maxPlayers int
		var mapName string
		var startedFlag int
		var finishedFlag int
		var activePlayer sql.NullString
		err := rows.Scan(&id, &name, &ownerUserId, &minPlayers, &maxPlayers, &mapName, &startedFlag, &finishedFlag, &activePlayer)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch game row: %v", err)
		}

		owner, err := server.getUserById(ownerUserId)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch owner: %v", err)
		}

		joinedUserIds, err := server.getJoinedUsers(id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch joined users: %v", err)
		}
		joinedUsers := make([]*User, 0, len(joinedUserIds))
		for userId := range joinedUserIds {
			user, err := server.getUserById(userId)
			if err != nil {
				return nil, fmt.Errorf("failed to get user %s: %v", userId, err)
			}
			joinedUsers = append(joinedUsers, user)
		}

		games = append(games, &GameSummary{
			Id:           id,
			Name:         name,
			Started:      startedFlag != 0,
			Finished:     finishedFlag != 0,
			MinPlayers:   minPlayers,
			MaxPlayers:   maxPlayers,
			MapName:      mapName,
			ActivePlayer: activePlayer.String,
			OwnerUser:    owner,
			JoinedUsers:  joinedUsers,
		})
	}

	return games, nil
}

type GetMyGamesRequest struct {
}
type GetMyGamesResponse struct {
	Games []*GameSummary `json:"games"`
}

func (server *GameServer) getMyGames(ctx *RequestContext, req *GetMyGamesRequest) (resp *GetMyGamesResponse, err error) {
	stmt, err := server.db.Prepare("SELECT G.id,G.name,G.owner_user_id,G.min_players,G.max_players,G.map_name,G.started,G.finished,G.active_player_id FROM games G" +
		" INNER JOIN game_player_map ON game_player_map.game_id=G.id WHERE game_player_map.player_user_id=?" +
		" ORDER by G.created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(ctx.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()
	games, err := server.getGamesSummaries(rows)
	if err != nil {
		return nil, err
	}

	return &GetMyGamesResponse{
		Games: games,
	}, nil
}

type CustomColors struct {
	PlayerColors []string `json:"playerColors"`
	GoodsColors  []string `json:"goodsColors"`
}

type GetMyProfileRequest struct {
}
type GetMyProfileResponse struct {
	Id                        string        `json:"id"`
	Nickname                  string        `json:"nickname"`
	Email                     string        `json:"email"`
	GoogleId                  string        `json:"googleId"`
	DiscordId                 string        `json:"discordId"`
	EmailNotificationsEnabled bool          `json:"emailNotificationsEnabled"`
	DiscordTurnAlertsEnabled  bool          `json:"discordTurnAlertsEnabled"`
	ColorPreferences          []int         `json:"colorPreferences"`
	CustomColors              *CustomColors `json:"customColors"`
	Webhooks                  []string      `json:"webhooks"`
}

func (server *GameServer) getMyProfile(ctx *RequestContext, req *GetMyGamesRequest) (resp *GetMyProfileResponse, err error) {
	stmt, err := server.db.Prepare("SELECT nickname,email,discord_user_id,google_user_id,email_notifications_enabled,discord_turn_alerts_enabled,color_preferences,custom_colors,webhooks FROM users WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(ctx.User.Id)
	var nickname string
	var email sql.NullString
	var discordId sql.NullString
	var googleId sql.NullString
	var emailNotificationsEnabled int
	var discordTurnAlertsEnabled int
	var colorPreferencesStr sql.NullString
	var customColorsStr sql.NullString
	var webhooksStr sql.NullString
	err = row.Scan(&nickname, &email, &discordId, &googleId, &emailNotificationsEnabled, &discordTurnAlertsEnabled, &colorPreferencesStr, &customColorsStr, &webhooksStr)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %v", err)
	}

	var colorPreferences []int
	if colorPreferencesStr.Valid {
		err = json.Unmarshal([]byte(colorPreferencesStr.String), &colorPreferences)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal color preferences: %v", err)
		}
	}
	var customColors *CustomColors
	if customColorsStr.Valid {
		customColors = new(CustomColors)
		err = json.Unmarshal([]byte(customColorsStr.String), customColors)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom colors: %v", err)
		}
	}

	var webhooks []string
	if webhooksStr.Valid {
		err = json.Unmarshal([]byte(webhooksStr.String), &webhooks)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhooks: %v", err)
		}
	}

	return &GetMyProfileResponse{
		Id:                        ctx.User.Id,
		Nickname:                  nickname,
		Email:                     email.String,
		DiscordId:                 discordId.String,
		GoogleId:                  googleId.String,
		EmailNotificationsEnabled: emailNotificationsEnabled != 0,
		DiscordTurnAlertsEnabled:  discordTurnAlertsEnabled != 0,
		ColorPreferences:          colorPreferences,
		CustomColors:              customColors,
		Webhooks:                  webhooks,
	}, nil
}

type SetMyProfileRequest struct {
	EmailNotificationsEnabled bool          `json:"emailNotificationsEnabled"`
	DiscordTurnAlertsEnabled  bool          `json:"discordTurnAlertsEnabled"`
	ColorPreferences          []int         `json:"colorPreferences"`
	CustomColors              *CustomColors `json:"customColors"`
	Webhooks                  []string      `json:"webhooks"`
}
type SetMyProfileResponse struct {
}

func (server *GameServer) setMyProfile(ctx *RequestContext, req *SetMyProfileRequest) (resp *SetMyProfileResponse, err error) {
	stmt, err := server.db.Prepare("UPDATE users SET email_notifications_enabled=?,color_preferences=?,custom_colors=?,webhooks=?,discord_turn_alerts_enabled=? WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	var colorPreferencesStr sql.NullString
	if len(req.ColorPreferences) != 0 {
		jsonBytes, err := json.Marshal(req.ColorPreferences)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal color preferences: %v", err)
		}
		colorPreferencesStr.Valid = true
		colorPreferencesStr.String = string(jsonBytes)
	}

	var customColorsStr sql.NullString
	if req.CustomColors != nil {
		jsonBytes, err := json.Marshal(req.CustomColors)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom colors: %v", err)
		}
		customColorsStr.Valid = true
		customColorsStr.String = string(jsonBytes)
	}

	var emailNotificationsEnabled int
	if req.EmailNotificationsEnabled {
		emailNotificationsEnabled = 1
	}
	var discordTurnAlertsEnabled int
	if req.DiscordTurnAlertsEnabled {
		discordTurnAlertsEnabled = 1
	}

	var webhooksStr sql.NullString
	if len(req.Webhooks) != 0 {
		jsonBytes, err := json.Marshal(req.Webhooks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal webhooks: %v", err)
		}
		webhooksStr.Valid = true
		webhooksStr.String = string(jsonBytes)
	}

	_, err = stmt.Exec(emailNotificationsEnabled, colorPreferencesStr, customColorsStr, webhooksStr, discordTurnAlertsEnabled, ctx.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	return &SetMyProfileResponse{}, nil
}

type GetGameChatRequest struct {
	GameId string `json:"gameId"`
	// Get all messages sent strictly after this time, in unix epoch seconds
	After int `json:"after"`
}

type GameChatMessage struct {
	UserId    string `json:"userId"`
	Timestamp int    `json:"timestamp"`
	Message   string `json:"message"`
}

type GetGameChatResponse struct {
	Messages []*GameChatMessage `json:"messages"`
}

func (server *GameServer) getGameChat(ctx *RequestContext, req *GetGameChatRequest) (resp *GetGameChatResponse, err error) {
	if req.GameId == "" {
		return nil, &api.HttpError{"missing gameId parameter", http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("SELECT timestamp,user_id,message FROM game_chat WHERE game_id=? AND timestamp > ? ORDER BY timestamp ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(req.GameId, req.After)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var messages []*GameChatMessage
	for rows.Next() {
		var timestamp int
		var userId string
		var message string
		err = rows.Scan(&timestamp, &userId, &message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		messages = append(messages, &GameChatMessage{
			Timestamp: timestamp,
			UserId:    userId,
			Message:   message,
		})
	}

	return &GetGameChatResponse{
		Messages: messages,
	}, nil
}

type SendGameChatRequest struct {
	GameId  string `json:"gameId"`
	Message string `json:"message"`
}
type SendGameChatResponse struct {
}

func (server *GameServer) sendGameChat(ctx *RequestContext, req *SendGameChatRequest) (resp *SendGameChatResponse, err error) {
	if req.GameId == "" {
		return nil, &api.HttpError{"missing gameId parameter", http.StatusBadRequest}
	}
	if req.Message == "" {
		return nil, &api.HttpError{"missing message parameter", http.StatusBadRequest}
	}
	users, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if _, ok := users[ctx.User.Id]; !ok {
		return nil, &api.HttpError{"can only send messages in games you are in", http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("INSERT INTO game_chat (game_id, timestamp, user_id, message) VALUES(?, ?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(req.GameId, time.Now().Unix(), ctx.User.Id, req.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to add row: %v", err)
	}

	return &SendGameChatResponse{}, nil
}

type PollGameStatusRequest struct {
	GameId string `json:"gameId"`
}
type PollGameStatusResponse struct {
	LastMove int `json:"lastMove"`
	LastChat int `json:"lastChat"`
}

func (server *GameServer) pollGameStatus(ctx *RequestContext, req *PollGameStatusRequest) (resp *PollGameStatusResponse, err error) {
	if req.GameId == "" {
		return nil, &api.HttpError{"missing gameId parameter", http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("SELECT MAX(timestamp) FROM game_log WHERE game_id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)

	var lastMove sql.NullInt64
	err = row.Scan(&lastMove)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	stmt, err = server.db.Prepare("SELECT MAX(timestamp) FROM game_chat WHERE game_id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	row = stmt.QueryRow(req.GameId)

	var lastChat sql.NullInt64
	err = row.Scan(&lastChat)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	return &PollGameStatusResponse{
		LastMove: int(lastMove.Int64),
		LastChat: int(lastChat.Int64),
	}, nil
}

type UndoMoveRequest struct {
	GameId string `json:"gameId"`
}
type UndoMoveResponse struct {
}

func (server *GameServer) undoMove(ctx *RequestContext, req *UndoMoveRequest) (resp *UndoMoveResponse, err error) {
	trx, err := server.db.Begin()
	if err != nil {
		return nil, err
	}
	defer trx.Rollback()

	var lastTimestamp int
	var lastPlayer string
	var lastWasReversible int
	stmt, err := trx.Prepare("SELECT timestamp,user_id,reversible FROM game_log WHERE game_id=? ORDER BY timestamp DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(req.GameId)
	err = row.Scan(&lastTimestamp, &lastPlayer, &lastWasReversible)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, err
	}
	if lastPlayer != ctx.User.Id {
		return nil, &api.HttpError{"caller did not perform the last move", http.StatusBadRequest}
	}
	if lastWasReversible == 0 {
		return nil, &api.HttpError{"last move is not reversible", http.StatusBadRequest}
	}

	var priorState string
	stmt, err = trx.Prepare("SELECT new_game_state FROM game_log WHERE game_id=? AND timestamp < ? ORDER BY TIMESTAMP DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	row = stmt.QueryRow(req.GameId, lastTimestamp)
	err = row.Scan(&priorState)
	if err != nil {
		return nil, err
	}

	// Apply the reverted game state
	stmt, err = trx.Prepare("UPDATE games SET active_player_id=?,game_state=? WHERE id=?")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec(lastPlayer, priorState, req.GameId)
	if err != nil {
		return nil, err
	}

	// Add a log of the undo action
	stmt, err = trx.Prepare("INSERT INTO game_log (game_id,timestamp,user_id,action,description,new_active_player,new_game_state,reversible) VALUES(?, ?, ?, 'undo', ?, ?, ?, 0)")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec(req.GameId, time.Now().Unix(), lastPlayer, fmt.Sprintf("%s undid their previous action", ctx.User.Nickname),
		lastPlayer, priorState)
	if err != nil {
		return nil, err
	}
	err = trx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &UndoMoveResponse{}, nil
}
