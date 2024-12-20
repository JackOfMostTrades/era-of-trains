package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"time"
)

type GameServer struct {
	config *Config
	db     *sql.DB
	maps   map[string]*BasicMap
}

type User struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Id       string `json:"id"`
}

type LoginRequest struct {
	AccessToken string `json:"accessToken"`
	// For development purpsose only. If devmode is enabled, sign-in with the given nickname, bypassing actual authentication
	DevNickname string `json:"devNickname"`
}

type LoginResponse struct {
}

func (server *GameServer) Close() error {
	err := server.db.Close()
	if err != nil {
		return err
	}
	return nil
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
				return nil, &HttpError{fmt.Sprintf("invalid nickname: %s", req.DevNickname), http.StatusBadRequest}
			}
			return nil, fmt.Errorf("failed to lookup user: %v", err)
		}
	} else {
		if req.AccessToken == "" {
			return nil, &HttpError{"Missing access token", http.StatusBadRequest}
		}

		userInfoReq, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo", nil)
		if err != nil {
			return nil, err
		}
		userInfoReq.Header.Add("Authorization", "Bearer "+req.AccessToken)
		userInfoRes, err := http.DefaultClient.Do(userInfoReq)
		if err != nil {
			return nil, err
		}
		if userInfoRes.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get user info from access token: %d", userInfoRes.StatusCode)
		}
		defer userInfoRes.Body.Close()

		userInfoResponse := struct {
			Id            string `json:"id"`
			Email         string `json:"email"`
			VerifiedEmail bool   `json:"verified_email"`
			Picture       string `json:"picture"`
		}{}
		err = json.NewDecoder(userInfoRes.Body).Decode(&userInfoResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to decode userinfo response: %v", err)
		}

		googleUserId := userInfoResponse.Id
		if googleUserId == "" {
			return nil, fmt.Errorf("failed to get user id from user info")
		}

		stmt, err := server.db.Prepare("SELECT id FROM users WHERE google_user_id=?")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %v", err)
		}
		defer stmt.Close()

		row := stmt.QueryRow(googleUserId)
		if err != nil {
			return nil, fmt.Errorf("failed to excute statement: %v", err)
		}

		err = row.Scan(&userId)
		if err != nil {
			if err == sql.ErrNoRows {
				// Fallback to finding a user row with matching email and no user ID
				stmt, err := server.db.Prepare("SELECT id FROM users WHERE email=? AND google_user_id IS NULL")
				if err != nil {
					if err != nil {
						return nil, fmt.Errorf("failed to excute statement: %v", err)
					}
				}
				defer stmt.Close()
				row = stmt.QueryRow(userInfoResponse.Email)
				err = row.Scan(&userId)
				if err != nil && err != sql.ErrNoRows {
					if err != sql.ErrNoRows {
						return nil, fmt.Errorf("failed to excute statement: %v", err)
					}
					return nil, &HttpError{fmt.Sprintf("google user is not registered (%s / %s)", googleUserId, userInfoResponse.Email), http.StatusPreconditionFailed}
				}

				stmt, err = server.db.Prepare("UPDATE users SET google_user_id=? WHERE id=?")
				if err != nil {
					return nil, fmt.Errorf("failed to excute statement: %v", err)
				}
				defer stmt.Close()
				_, err = stmt.Exec(googleUserId, userId)
				if err != nil {
					return nil, fmt.Errorf("failed to excute statement: %v", err)
				}
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
	})
	return &LoginResponse{}, nil
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
		Expires:  time.Unix(0, 0),
	})
	return &LogoutResponse{}, nil
}

type CreateGameRequest struct {
	Name       string `json:"name"`
	NumPlayers int    `json:"numPlayers"`
	MapName    string `json:"mapName"`
}

type CreateGameResponse struct {
	Id string `json:"id"`
}

func (server *GameServer) createGame(ctx *RequestContext, req *CreateGameRequest) (resp *CreateGameResponse, err error) {
	stmt, err := server.db.Prepare("INSERT INTO games (id,created_at,name,num_players,map_name,owner_user_id,started,finished) VALUES (?,?,?,?,?,?,0,0)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %v", err)
	}
	_, err = stmt.Exec(id.String(), time.Now().Unix(), req.Name, req.NumPlayers, req.MapName, ctx.User.Id)
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
	stmt, err := server.db.Prepare("SELECT num_players,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var numPlayers int
	var startedFlag int
	err = row.Scan(&numPlayers, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &HttpError{"game has already started", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if _, ok := joinedUsers[ctx.User.Id]; ok {
		return nil, &HttpError{"you have already joined this game", http.StatusBadRequest}
	}
	if len(joinedUsers) >= numPlayers {
		return nil, &HttpError{"this game is already full", http.StatusBadRequest}
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

	return &JoinGameResponse{}, nil
}

type LeaveGameRequest struct {
	GameId string `json:"gameId"`
}
type LeaveGameResponse struct {
}

func (server *GameServer) leaveGame(ctx *RequestContext, req *LeaveGameRequest) (resp *LeaveGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,num_players,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var numPlayers int
	var startedFlag int
	err = row.Scan(&ownerUserId, &numPlayers, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &HttpError{"game has already started", http.StatusBadRequest}
	}
	if ownerUserId == ctx.User.Id {
		return nil, &HttpError{"you cannot leave a game that you created", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if _, ok := joinedUsers[ctx.User.Id]; !ok {
		return nil, &HttpError{"you are not joined to this game", http.StatusBadRequest}
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

	return &LeaveGameResponse{}, nil
}

type StartGameRequest struct {
	GameId string `json:"gameId"`
}
type StartGameResponse struct {
}

func (server *GameServer) startGame(ctx *RequestContext, req *StartGameRequest) (resp *StartGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,num_players,map_name,started FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var numPlayers int
	var mapName string
	var startedFlag int
	err = row.Scan(&ownerUserId, &numPlayers, &mapName, &startedFlag)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag != 0 {
		return nil, &HttpError{"game has already started", http.StatusBadRequest}
	}
	if ownerUserId != ctx.User.Id {
		return nil, &HttpError{"you are not the owner of this game", http.StatusBadRequest}
	}

	joinedUsers, err := server.getJoinedUsers(req.GameId)
	if err != nil {
		return nil, err
	}

	if len(joinedUsers) != numPlayers {
		return nil, &HttpError{"game is not full", http.StatusBadRequest}
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
	shuffledColors := make([]int, 6) // 6 colors available to players, regardless of player count
	for idx := 0; idx < len(shuffledColors); idx++ {
		shuffledColors[idx] = idx
	}
	rand.Shuffle(len(shuffledColors), func(i, j int) {
		shuffledColors[i], shuffledColors[j] = shuffledColors[j], shuffledColors[i]
	})
	playerColor := make(map[string]int)
	for idx, playerId := range playerOrder {
		playerColor[playerId] = shuffledColors[idx]
	}

	// Setup initial game state
	gameState := &GameState{
		ActivePlayer:      playerOrder[0],
		PlayerOrder:       playerOrder,
		PlayerColor:       playerColor,
		PlayerShares:      make(map[string]int),
		PlayerLoco:        make(map[string]int),
		PlayerIncome:      make(map[string]int),
		PlayerActions:     make(map[string]SpecialAction),
		PlayerCash:        make(map[string]int),
		AuctionState:      make(map[string]int),
		GamePhase:         SHARES_GAME_PHASE,
		TurnNumber:        1,
		MovingGoodsRound:  0,
		PlayerHasDoneLoco: make(map[string]bool),
		Links:             nil,
		Urbanizations:     nil,
		CubeBag: map[Color]int{
			BLACK:  16,
			RED:    20,
			YELLOW: 20,
			BLUE:   20,
			PURPLE: 20,
		},
		Cubes:           nil,
		GoodsGrowth:     make([][]Color, 20),
		ProductionCubes: nil,
	}
	for _, userId := range playerOrder {
		gameState.PlayerShares[userId] = 2
		gameState.PlayerLoco[userId] = 1
		gameState.PlayerIncome[userId] = 0
		gameState.PlayerCash[userId] = 10
	}

	// Populate the goods growth table
	for i := 0; i < 12; i++ {
		gameState.GoodsGrowth[i] = make([]Color, 3)
		for j := 0; j < 3; j++ {
			cube, err := gameState.drawCube()
			if err != nil {
				return nil, fmt.Errorf("failed to draw cube: %v", err)
			}
			gameState.GoodsGrowth[i][j] = cube
		}
	}
	for i := 12; i < 20; i++ {
		gameState.GoodsGrowth[i] = make([]Color, 2)
		for j := 0; j < 2; j++ {
			cube, err := gameState.drawCube()
			if err != nil {
				return nil, fmt.Errorf("failed to draw cube: %v", err)
			}
			gameState.GoodsGrowth[i][j] = cube
		}
	}

	theMap := server.maps[mapName]
	if theMap == nil {
		return nil, fmt.Errorf("failed to lookup map: %s", mapName)
	}
	for _, startingCubeSpec := range theMap.StartingCubes {
		for i := 0; i < startingCubeSpec.Number; i++ {
			cube, err := gameState.drawCube()
			if err != nil {
				return nil, fmt.Errorf("failed to draw cube: %v", err)
			}
			gameState.Cubes = append(gameState.Cubes, &BoardCube{
				Color: cube,
				Hex:   startingCubeSpec.Coordinate,
			})
		}
	}

	gameStateStr, err := json.Marshal(gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game state: %v", err)
	}
	stmt, err = server.db.Prepare("UPDATE games SET game_state=? WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(string(gameStateStr), req.GameId)
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
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Started     bool       `json:"started"`
	Finished    bool       `json:"finished"`
	NumPlayers  int        `json:"numPlayers"`
	MapName     string     `json:"mapName"`
	OwnerUser   *User      `json:"ownerUser"`
	JoinedUsers []*User    `json:"joinedUsers"`
	GameState   *GameState `json:"gameState"`
}

func (server *GameServer) viewGame(ctx *RequestContext, req *ViewGameRequest) (resp *ViewGameResponse, err error) {
	stmt, err := server.db.Prepare("SELECT name,owner_user_id,num_players,map_name,started,finished,game_state FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(req.GameId)
	var name string
	var ownerUserId string
	var numPlayers int
	var mapName string
	var startedFlag int
	var finishedFlag int
	var gameStateStr sql.NullString
	err = row.Scan(&name, &ownerUserId, &numPlayers, &mapName, &startedFlag, &finishedFlag, &gameStateStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
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

	var gameState *GameState
	if gameStateStr.Valid {
		gameState = new(GameState)
		err = json.Unmarshal([]byte(gameStateStr.String), gameState)
		if err != nil {
			return nil, fmt.Errorf("failed to parse game state: %v", err)
		}
	}

	res := &ViewGameResponse{
		Id:          req.GameId,
		Name:        name,
		Started:     startedFlag != 0,
		Finished:    finishedFlag != 0,
		NumPlayers:  numPlayers,
		MapName:     mapName,
		OwnerUser:   owner,
		JoinedUsers: joinedUsers,
		GameState:   gameState,
	}

	return res, nil
}

type GameSummary struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Started     bool    `json:"started"`
	Finished    bool    `json:"finished"`
	NumPlayers  int     `json:"numPlayers"`
	MapName     string  `json:"mapName"`
	OwnerUser   *User   `json:"ownerUser"`
	JoinedUsers []*User `json:"joinedUsers"`
}

type ListGamesRequest struct {
}
type ListGamesResponse struct {
	Games []*GameSummary `json:"games"`
}

func (server *GameServer) listGames(ctx *RequestContext, req *ListGamesRequest) (resp *ListGamesResponse, err error) {
	stmt, err := server.db.Prepare("SELECT id,name,owner_user_id,num_players,map_name,started,finished FROM games ORDER by created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	var games []*GameSummary

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var name string
		var ownerUserId string
		var numPlayers int
		var mapName string
		var startedFlag int
		var finishedFlag int
		err = rows.Scan(&id, &name, &ownerUserId, &numPlayers, &mapName, &startedFlag, &finishedFlag)
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
			Id:          id,
			Name:        name,
			Started:     startedFlag != 0,
			Finished:    finishedFlag != 0,
			NumPlayers:  numPlayers,
			MapName:     mapName,
			OwnerUser:   owner,
			JoinedUsers: joinedUsers,
		})
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
}

type GetGameLogsRequest struct {
	GameId string `json:"gameId"`
}
type GetGameLogsResponse struct {
	Logs []*GameLogEntry `json:"logs"`
}

func (server *GameServer) getGameLogs(ctx *RequestContext, req *GetGameLogsRequest) (resp *GetGameLogsResponse, err error) {
	stmt, err := server.db.Prepare("SELECT timestamp,user_id,action,description FROM game_log WHERE game_id=? ORDER BY timestamp ASC")
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
		var action string
		var description string
		err = rows.Scan(&timestamp, &userId, &action, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		entries = append(entries, &GameLogEntry{
			Timestamp:   timestamp,
			UserId:      userId,
			Action:      action,
			Description: description,
		})
	}

	return &GetGameLogsResponse{
		Logs: entries,
	}, nil
}
