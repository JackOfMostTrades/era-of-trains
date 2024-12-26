package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"net/http"
	"slices"
	"strings"
	"time"
)

type ActionName string

const (
	SharesActionName       ActionName = "shares"
	BidActionName          ActionName = "bid"
	ChooseActionName       ActionName = "choose_action"
	BuildActionName        ActionName = "build"
	MoveGoodsActionName    ActionName = "move_goods"
	ProduceGoodsActionName ActionName = "produce_goods"
)

type SharesAction struct {
	Amount int `json:"amount"`
}

type BidAction struct {
	Amount int `json:"amount"`
}

type ChooseAction struct {
	Action common.SpecialAction `json:"action"`
}

type TownPlacement struct {
	Track common.Direction  `json:"track"`
	Hex   common.Coordinate `json:"hex"`
}

type TrackRedirect struct {
	Track common.Direction  `json:"track"`
	Hex   common.Coordinate `json:"hex"`
}

type TrackPlacement struct {
	Track [2]common.Direction `json:"track""`
	Hex   common.Coordinate   `json:"hex"`
}

type InterurbanLinkPlacement struct {
	Track common.Direction  `json:"track"`
	Hex   common.Coordinate `json:"hex"`
}

type BuildAction struct {
	Urbanization             *common.Urbanization       `json:"urbanization"`
	TownPlacements           []*TownPlacement           `json:"townPlacements"`
	TrackRedirects           []*TrackRedirect           `json:"trackRedirects"`
	TrackPlacements          []*TrackPlacement          `json:"trackPlacements"`
	InterurbanLinkPlacements []*InterurbanLinkPlacement `json:"interurbanLinkPlacements"`
}

type MoveGoodsAction struct {
	StartingLocation common.Coordinate  `json:"startingLocation"`
	Color            common.Color       `json:"color"`
	Path             []common.Direction `json:"path"`
	Loco             bool               `json:"loco"`
}

type ProduceGoodsAction struct {
	// List (corresponding the cubes in the same order as ProductionCubes in the game state) with X,Y coordinates
	// corresponding to which city (X) and which spot (Y) within that city
	Destinations []common.Coordinate `json:"destinations"`
}

type ConfirmMoveRequest struct {
	GameId             string              `json:"gameId"`
	ActionName         ActionName          `json:"actionName"`
	SharesAction       *SharesAction       `json:"sharesAction"`
	BidAction          *BidAction          `json:"bidAction"`
	ChooseAction       *ChooseAction       `json:"chooseAction"`
	BuildAction        *BuildAction        `json:"buildAction"`
	MoveGoodsAction    *MoveGoodsAction    `json:"moveGoodsAction"`
	ProduceGoodsAction *ProduceGoodsAction `json:"produceGoodsAction"`
}
type ConfirmMoveResponse struct {
}

type confirmMoveHandler struct {
	gameMap        maps.GameMap
	gameState      *common.GameState
	activePlayer   string
	logs           []string
	playerIdToNick map[string]string
	randProvider   common.RandProvider
}

func newConfirmMoveHandler(server *GameServer, gameId string, gameMap maps.GameMap, gameState *common.GameState, activePlayer string) (*confirmMoveHandler, error) {
	handler := &confirmMoveHandler{
		gameMap:        gameMap,
		gameState:      gameState,
		activePlayer:   activePlayer,
		playerIdToNick: make(map[string]string),
		randProvider:   server.randProvider,
	}

	stmt, err := server.db.Prepare("SELECT id,nickname FROM users INNER JOIN game_player_map ON users.id=game_player_map.player_user_id WHERE game_player_map.game_id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var userId string
		var nickname string
		err = rows.Scan(&userId, &nickname)
		if err != nil {
			return nil, fmt.Errorf("failed to read query result: %v", err)
		}
		handler.playerIdToNick[userId] = nickname
	}

	return handler, nil
}

func (handler *confirmMoveHandler) Log(format string, a ...any) {
	handler.logs = append(handler.logs, fmt.Sprintf(format, a...))
}

func (handler *confirmMoveHandler) PlayerNick(playerId string) string {
	if nick, ok := handler.playerIdToNick[playerId]; ok {
		return nick
	}
	return "<unknown: " + playerId + ">"
}

func (handler *confirmMoveHandler) ActivePlayerNick() string {
	return handler.PlayerNick(handler.activePlayer)
}

func (server *GameServer) confirmMove(ctx *RequestContext, req *ConfirmMoveRequest) (resp *ConfirmMoveResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,num_players,map_name,started,finished,game_state,active_player_id FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var numPlayers int
	var mapName string
	var startedFlag int
	var finishedFlag int
	var gameStateStr string
	var activePlayer string
	err = row.Scan(&ownerUserId, &numPlayers, &mapName, &startedFlag, &finishedFlag, &gameStateStr, &activePlayer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag == 0 {
		return nil, &HttpError{fmt.Sprintf("cannot make a move if game hasn't started yet: %s", req.GameId), http.StatusBadRequest}
	}
	if finishedFlag != 0 {
		return nil, &HttpError{fmt.Sprintf("cannot make a move if game has finished: %s", req.GameId), http.StatusBadRequest}
	}

	gameState := new(common.GameState)
	err = json.Unmarshal([]byte(gameStateStr), gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to parse game state: %v", err)
	}
	if activePlayer != ctx.User.Id {
		return nil, &HttpError{fmt.Sprintf("user [%s] is not the active player [%s]", ctx.User.Id, activePlayer), http.StatusPreconditionFailed}
	}

	gameMap := server.gameMaps[mapName]

	handler, err := newConfirmMoveHandler(server, req.GameId, gameMap, gameState, activePlayer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize handler: %v", err)
	}

	err = handler.handleAction(req)
	if err != nil {
		return nil, err
	}

	newGameStateStr, err := json.Marshal(gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game state: %v", err)
	}

	// Determine if the game is over
	turnLimit := 10
	if len(gameState.PlayerOrder) == 6 {
		turnLimit = 6
	} else if len(gameState.PlayerOrder) == 5 {
		turnLimit = 7
	} else if len(gameState.PlayerOrder) == 4 {
		turnLimit = 8
	}
	if gameState.TurnNumber > turnLimit {
		finishedFlag = 1
	}

	// Log the action
	stmt, err = server.db.Prepare("INSERT INTO game_log (game_id,timestamp,user_id,action,description,new_active_player,new_game_state) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	reqString, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialze the request for logging: %v", err)
	}
	_, err = stmt.Exec(req.GameId, time.Now().Unix(), ctx.User.Id, string(reqString), strings.Join(handler.logs, "\n"),
		handler.activePlayer, string(newGameStateStr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Update the game state
	stmt, err = server.db.Prepare("UPDATE games SET active_player_id=?,game_state=?,finished=? WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(handler.activePlayer, string(newGameStateStr), finishedFlag, req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Send notifications
	if finishedFlag != 0 {
		userIds, err := server.getJoinedUsers(req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to get game users: %v", err)
		}
		for userId := range userIds {
			err = server.notifyPlayer(req.GameId, userId)
			if err != nil {
				return nil, fmt.Errorf("failed to notify user of game end: %v", err)
			}
		}
	} else {
		// Notify the next player that it is their turn if the active player changed
		if handler.activePlayer != ctx.User.Id {
			err = server.notifyPlayer(req.GameId, handler.activePlayer)
			if err != nil {
				return nil, fmt.Errorf("failed to notify user it's their turn: %v", err)
			}
		}
	}

	return &ConfirmMoveResponse{}, nil
}

func (handler *confirmMoveHandler) handleAction(req *ConfirmMoveRequest) error {
	var err error
	switch req.ActionName {
	case SharesActionName:
		err = handler.handleSharesAction(req.SharesAction)
	case BidActionName:
		err = handler.handleBidAction(req.BidAction)
	case ChooseActionName:
		err = handler.handleChooseAction(req.ChooseAction)
	case BuildActionName:
		err = handler.handleBuildAction(req.BuildAction)
	case MoveGoodsActionName:
		err = handler.handleMoveGoodsAction(req.MoveGoodsAction)
	case ProduceGoodsActionName:
		err = handler.handleProduceGoodsAction(req.ProduceGoodsAction)
	default:
		err = &HttpError{fmt.Sprintf("invalid action: %s", req.ActionName), http.StatusBadRequest}
	}
	if err != nil {
		return err
	}
	return nil
}

func (handler *confirmMoveHandler) handleSharesAction(sharesAction *SharesAction) error {
	gameState := handler.gameState
	if sharesAction == nil || sharesAction.Amount < 0 {
		return &HttpError{"missing shares action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.SHARES_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := handler.activePlayer
	newSharesCount := gameState.PlayerShares[currentPlayer] + sharesAction.Amount
	if newSharesCount > 15 {
		return &HttpError{"cannot take more than 15 shares", http.StatusBadRequest}
	}
	gameState.PlayerShares[currentPlayer] = newSharesCount
	gameState.PlayerCash[currentPlayer] += 5 * sharesAction.Amount

	handler.Log("%s takes %d shares.", handler.ActivePlayerNick(), sharesAction.Amount)

	nextPlayerId := ""
	for i := 0; i < len(gameState.PlayerOrder)-1; i++ {
		if gameState.PlayerOrder[i] == currentPlayer {
			nextPlayerId = gameState.PlayerOrder[i+1]
			break
		}
	}
	if nextPlayerId == "" {
		// Advance game phase
		gameState.GamePhase = common.AUCTION_GAME_PHASE
		handler.activePlayer = gameState.PlayerOrder[0]
	} else {
		handler.activePlayer = nextPlayerId
	}

	return nil
}

func (handler *confirmMoveHandler) handleBidAction(bidAction *BidAction) error {
	gameState := handler.gameState
	if bidAction == nil {
		return &HttpError{"missing bid action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.AUCTION_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := handler.activePlayer

	gotoNextPhase := false
	// If the user is passing
	if bidAction.Amount < 0 {
		// How many users have already passed
		passCount := 0
		for _, bidAmount := range gameState.AuctionState {
			if bidAmount < 0 {
				passCount += 1
			}
		}

		lastBid := gameState.AuctionState[currentPlayer]
		var cashToPay int
		if passCount == 0 {
			// Last player does not pay
			cashToPay = 0
		} else if (len(gameState.PlayerOrder) - passCount) > 2 {
			// In the middle, pay half price (rounded up)
			cashToPay = lastBid/2 + (lastBid % 2)
		} else {
			// Only two players left to pass, pay full price of the last bid
			cashToPay = lastBid
		}

		gameState.PlayerCash[currentPlayer] -= cashToPay
		// Set auction state to pass order (-1 first to pass, -2 second to pass, etc)
		gameState.AuctionState[currentPlayer] = (-1 * passCount) - 1

		handler.Log("%s passes, becoming player number %d and paying $%d based on their bid of $%d.", handler.ActivePlayerNick(),
			len(gameState.PlayerOrder)-passCount, cashToPay, lastBid)

		// If all but one other play has passed, that player implicitly passes since there's no one else left
		if passCount == len(gameState.PlayerOrder)-2 {
			// Implicitly pass the remaining player
			for _, playerId := range gameState.PlayerOrder {
				if bidAmount := gameState.AuctionState[playerId]; bidAmount >= 0 {
					gameState.PlayerCash[playerId] -= bidAmount
					gameState.AuctionState[playerId] = (-1 * passCount) - 2

					handler.Log("%s becomes first player as last player to not pass, and pays $%d.",
						handler.PlayerNick(playerId), bidAmount)
				}
			}
			gotoNextPhase = true
		}
		// FIXME: This shouldn't be reachable; it shouldn't be possible that there is only one player left who has to pass? TBD if this can be dropped
		if passCount == len(gameState.PlayerOrder)-1 {
			gotoNextPhase = true
		}

	} else if bidAction.Amount == 0 {
		// Bid amount of 0 indicates use of turn-order-pass
		handler.Log("%s uses turn-order pass.", handler.ActivePlayerNick())

		if gameState.PlayerActions[currentPlayer] != common.TURN_ORDER_PASS_SPECIAL_ACTION {
			return &HttpError{"current player cannot use turn-order pass", http.StatusBadRequest}
		}

		// Do not update this user's bid amount
		// Remove user's turn-order-pass action
		gameState.PlayerActions[currentPlayer] = ""

	} else {
		// User is increasing their bid
		playerCash := gameState.PlayerCash[currentPlayer]
		if bidAction.Amount >= playerCash {
			return &HttpError{fmt.Sprintf("bid amount [%d] greater than player's cash on hand %d", bidAction.Amount, playerCash), http.StatusBadRequest}
		}

		currentHighBid := 0
		for _, bidAmount := range gameState.AuctionState {
			if bidAmount > 0 && bidAmount > currentHighBid {
				currentHighBid = bidAmount
			}
		}
		if bidAction.Amount <= currentHighBid {
			return &HttpError{fmt.Sprintf("bid amount [%d] greater not higher than current high bid %d", bidAction.Amount, currentHighBid), http.StatusBadRequest}
		}

		// Update this user's bid
		gameState.AuctionState[currentPlayer] = bidAction.Amount

		handler.Log("%s bids $%d.", handler.ActivePlayerNick(), bidAction.Amount)
	}

	if gotoNextPhase {
		// Get the new player order from the auction state
		gameState.PlayerOrder = gameState.PlayerOrder[:len(gameState.AuctionState)]
		for userId, bidAmount := range gameState.AuctionState {
			// bidAmount of -1 should be first from end, bidAmount of -2 next from end, etc
			gameState.PlayerOrder[len(gameState.AuctionState)+bidAmount] = userId
		}
		// Then reset the auction state
		for userId := range gameState.AuctionState {
			delete(gameState.AuctionState, userId)
		}
		// Set the active player to the new first player
		handler.activePlayer = gameState.PlayerOrder[0]
		// Advance the game phase
		gameState.GamePhase = common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
		// Force-remove any chosen special actions as we advance into that phase
		for userId := range gameState.PlayerActions {
			gameState.PlayerActions[userId] = ""
		}

	} else {
		currentPlayerPosition := -1
		for idx, userId := range gameState.PlayerOrder {
			if userId == currentPlayer {
				currentPlayerPosition = idx
				break
			}
		}
		if currentPlayerPosition == -1 {
			return &HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
		}

		currentHighBid := 0
		for _, bidAmount := range gameState.AuctionState {
			if bidAmount > 0 && bidAmount > currentHighBid {
				currentHighBid = bidAmount
			}
		}

		nextPlayer := ""
		for i := 1; i < len(gameState.PlayerOrder); i++ {
			userId := gameState.PlayerOrder[(currentPlayerPosition+i)%len(gameState.PlayerOrder)]
			userBid := gameState.AuctionState[userId]
			// Users who have passed or have the current high bid do not go
			if userBid < 0 || userBid == currentHighBid {
				continue
			}
			nextPlayer = userId
			break
		}
		if nextPlayer == "" {
			return &HttpError{"unable to find next player", http.StatusInternalServerError}
		}
		handler.activePlayer = nextPlayer
	}

	return nil
}

func (handler *confirmMoveHandler) handleChooseAction(chooseAction *ChooseAction) error {
	gameState := handler.gameState
	if chooseAction == nil {
		return &HttpError{"missing choose action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	isValid := false
	for _, action := range common.ALL_SPECIAL_ACTIONS {
		if chooseAction.Action == action {
			isValid = true
			break
		}
	}
	if !isValid {
		return &HttpError{fmt.Sprintf("invalid action: %s", chooseAction.Action), http.StatusBadRequest}
	}

	isChosen := false
	for _, action := range gameState.PlayerActions {
		if action == chooseAction.Action {
			isChosen = true
			break
		}
	}
	if isChosen {
		return &HttpError{fmt.Sprintf("action has already been chosen: %s", chooseAction.Action), http.StatusBadRequest}
	}

	// Set the chosen action
	gameState.PlayerActions[handler.activePlayer] = chooseAction.Action
	handler.Log("%s chooses special action \"%s\".", handler.ActivePlayerNick(), chooseAction.Action)

	// Apply any immediate effects
	if chooseAction.Action == common.LOCO_SPECIAL_ACTION {
		if gameState.PlayerLoco[handler.activePlayer] < 6 {
			gameState.PlayerLoco[handler.activePlayer] += 1
			handler.Log("%s's loco increases to %d.", handler.ActivePlayerNick(), gameState.PlayerLoco[handler.activePlayer])
		}
	}

	// Advance current player
	currentPlayerPosition := -1
	for idx, userId := range gameState.PlayerOrder {
		if userId == handler.activePlayer {
			currentPlayerPosition = idx
			break
		}
	}
	if currentPlayerPosition == -1 {
		return &HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
	}
	// Advance the game phase if this was the last player
	if currentPlayerPosition == len(gameState.PlayerOrder)-1 {
		gameState.GamePhase = common.BUILDING_GAME_PHASE
		// If anyone has first build, they become the active player
		firstBuildUser := ""
		for userId, action := range gameState.PlayerActions {
			if action == common.FIRST_BUILD_SPECIAL_ACTION {
				firstBuildUser = userId
				break
			}
		}

		if firstBuildUser == "" {
			handler.activePlayer = gameState.PlayerOrder[0]
		} else {
			handler.activePlayer = firstBuildUser
		}
	} else {
		// Otherwise just advance the active player
		handler.activePlayer = gameState.PlayerOrder[currentPlayerPosition+1]
	}

	return nil
}

func (handler *confirmMoveHandler) handleBuildAction(buildAction *BuildAction) error {
	gameState := handler.gameState
	if buildAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.BUILDING_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	err := handler.performBuildAction(buildAction)
	if err != nil {
		return err
	}

	// If this was the first build player, just advance to the first player in normal order
	if gameState.PlayerActions[handler.activePlayer] == common.FIRST_BUILD_SPECIAL_ACTION {
		// If this was the first player anyway, go to next player
		if gameState.PlayerOrder[0] == handler.activePlayer {
			handler.activePlayer = gameState.PlayerOrder[1]
		} else {
			handler.activePlayer = gameState.PlayerOrder[0]
		}
	} else {
		// Otherwise just advance the player, skipping over a player if they have first build
		currentPlayerPosition := -1
		for idx, userId := range gameState.PlayerOrder {
			if userId == handler.activePlayer {
				currentPlayerPosition = idx
				break
			}
		}
		if currentPlayerPosition == -1 {
			return &HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
		}

		nextPlayer := ""
		for i := currentPlayerPosition + 1; i < len(gameState.PlayerOrder); i++ {
			player := gameState.PlayerOrder[i]
			if gameState.PlayerActions[player] == common.FIRST_BUILD_SPECIAL_ACTION {
				continue
			}
			nextPlayer = player
			break
		}

		// End of phase
		if nextPlayer == "" {
			gameState.GamePhase = common.MOVING_GOODS_GAME_PHASE
			gameState.MovingGoodsRound = 0
			firstMovePlayer := ""
			for userId, action := range gameState.PlayerActions {
				if action == common.FIRST_MOVE_SPECIAL_ACTION {
					firstMovePlayer = userId
					break
				}
			}

			if firstMovePlayer != "" {
				handler.activePlayer = firstMovePlayer
			} else {
				handler.activePlayer = gameState.PlayerOrder[0]
			}
		} else {
			handler.activePlayer = nextPlayer
		}
	}

	return nil
}

func (handler *confirmMoveHandler) handleMoveGoodsAction(moveGoodsAction *MoveGoodsAction) error {
	gameState := handler.gameState
	gameMap := handler.gameMap
	if moveGoodsAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.MOVING_GOODS_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	if moveGoodsAction.Loco {
		if gameState.PlayerHasDoneLoco[handler.activePlayer] {
			return &HttpError{"player has already done loco this phase", http.StatusBadRequest}
		}
		if gameState.PlayerLoco[handler.activePlayer] >= 6 {
			return &HttpError{"player's loco is already at max", http.StatusBadRequest}
		}
		gameState.PlayerHasDoneLoco[handler.activePlayer] = true
		gameState.PlayerLoco[handler.activePlayer] += 1
		handler.Log("%s skipped delivering a good and increased their loco to %d",
			handler.ActivePlayerNick(), gameState.PlayerLoco[handler.activePlayer])
	} else if moveGoodsAction.Color != common.NONE_COLOR {

		deliveryGraph := computeDeliveryGraph(gameState)

		// Verify that there is a cube on the board of a matching color and the start location
		foundCube := false
		for idx, boardCube := range gameState.Cubes {
			if boardCube.Color == moveGoodsAction.Color && boardCube.Hex == moveGoodsAction.StartingLocation {
				gameState.Cubes = DeleteFromSliceUnordered(idx, gameState.Cubes)
				foundCube = true
				break
			}
		}
		if !foundCube {
			return &HttpError{"no such cube", http.StatusBadRequest}
		}

		handler.Log("%s delivered a %s good cube from (%d,%d)",
			handler.ActivePlayerNick(), moveGoodsAction.Color.String(), moveGoodsAction.StartingLocation.X, moveGoodsAction.StartingLocation.Y)

		if len(moveGoodsAction.Path) > gameState.PlayerLoco[handler.activePlayer] {
			return &HttpError{"cannot move good further than current loca", http.StatusBadRequest}
		}

		loc := moveGoodsAction.StartingLocation
		seenCities := []common.Coordinate{loc}
		for idx, step := range moveGoodsAction.Path {
			if _, ok := deliveryGraph.hexToDirectionToLink[loc]; !ok {
				return &HttpError{"invalid path", http.StatusBadRequest}
			}
			if _, ok := deliveryGraph.hexToDirectionToLink[loc][step]; !ok {
				return &HttpError{"invalid path", http.StatusBadRequest}
			}

			link := deliveryGraph.hexToDirectionToLink[loc][step]
			loc = link.destination
			if slices.Index(seenCities, loc) != -1 {
				return &HttpError{"cannot repeat a city in the delivery path", http.StatusBadRequest}
			} else {
				seenCities = append(seenCities, loc)
			}

			hexType := gameMap.GetHexType(loc)
			var cityColor common.Color = common.NONE_COLOR
			if hexType == maps.TOWN_HEX_TYPE {
				var urbColor common.Color = common.NONE_COLOR
				for _, urb := range gameState.Urbanizations {
					if urb.Hex == loc {
						switch urb.City {
						case 0:
							urbColor = common.RED
						case 1:
							urbColor = common.BLUE
						case 2:
							urbColor = common.BLACK
						case 3:
							urbColor = common.BLACK
						case 4:
							urbColor = common.YELLOW
						case 5:
							urbColor = common.PURPLE
						case 6:
							urbColor = common.BLACK
						case 7:
							urbColor = common.BLACK
						}
						break
					}
				}
				cityColor = urbColor
			} else if hexType == maps.CITY_HEX_TYPE {
				cityColor = gameMap.GetCityColorForHex(gameState, loc)
			} else {
				return &HttpError{"invalid path", http.StatusBadRequest}
			}

			locBlocksCube := cityColor == moveGoodsAction.Color || gameMap.LocationBlocksCubePassage(moveGoodsAction.Color, loc)
			if idx != len(moveGoodsAction.Path)-1 && locBlocksCube {
				return &HttpError{"cannot pass through city matching the cube color", http.StatusBadRequest}
			}

			locAcceptsCube := cityColor == moveGoodsAction.Color || gameMap.LocationCanAcceptCube(moveGoodsAction.Color, loc)
			if idx == len(moveGoodsAction.Path)-1 && !locAcceptsCube {
				return &HttpError{"ending city must match cube color", http.StatusBadRequest}
			}

			if link.player != "" {
				gameState.PlayerIncome[link.player] += 1
			}

			handler.Log("The cube moved to (%d,%d) giving one income to %s", loc.X, loc.Y, handler.PlayerNick(link.player))
		}

		handler.Log("The cube finished its movement in (%d,%d)", loc.X, loc.Y)

		bonus := gameMap.GetDeliveryBonus(moveGoodsAction.Color)
		if bonus > 0 {
			gameState.PlayerIncome[handler.activePlayer] += 1
			handler.Log("%s receives an extra delivery bonus of %d income.", handler.ActivePlayerNick(), bonus)
		}

	} else {
		// Pass action, do nothing here
		handler.Log("%s skipped their move good action.", handler.ActivePlayerNick())
	}

	// If this was the first move player, just advance to the first player in normal order
	if gameState.PlayerActions[handler.activePlayer] == common.FIRST_MOVE_SPECIAL_ACTION {
		// If this was the first player anyway, go to next player
		if gameState.PlayerOrder[0] == handler.activePlayer {
			handler.activePlayer = gameState.PlayerOrder[1]
		} else {
			handler.activePlayer = gameState.PlayerOrder[0]
		}
	} else {
		// Otherwise just advance the player, skipping over a player if they have first move
		currentPlayerPosition := -1
		for idx, userId := range gameState.PlayerOrder {
			if userId == handler.activePlayer {
				currentPlayerPosition = idx
				break
			}
		}
		if currentPlayerPosition == -1 {
			return &HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
		}

		nextPlayer := ""
		for i := currentPlayerPosition + 1; i < len(gameState.PlayerOrder); i++ {
			player := gameState.PlayerOrder[i]
			if gameState.PlayerActions[player] == common.FIRST_MOVE_SPECIAL_ACTION {
				continue
			}
			nextPlayer = player
			break
		}

		// End of round
		if nextPlayer == "" {
			if gameState.MovingGoodsRound == 0 {
				gameState.MovingGoodsRound = 1

				firstMovePlayer := ""
				for userId, action := range gameState.PlayerActions {
					if action == common.FIRST_MOVE_SPECIAL_ACTION {
						firstMovePlayer = userId
						break
					}
				}

				if firstMovePlayer == "" {
					handler.activePlayer = gameState.PlayerOrder[0]
				} else {
					handler.activePlayer = firstMovePlayer
				}
			} else {
				// End of phase
				gameState.PlayerHasDoneLoco = make(map[string]bool)
				err := handler.executeIncomeAndExpenses()
				if err != nil {
					return err
				}
				gameState.GamePhase = common.GOODS_GROWTH_GAME_PHASE
				produceGoodsPlayer := ""
				for userId, action := range gameState.PlayerActions {
					if action == common.PRODUCTION_SPECIAL_ACTION {
						produceGoodsPlayer = userId
						break
					}
				}

				// Count the number of empty spaces on the goods growth chart
				emptyCount := 0
				for _, col := range gameState.GoodsGrowth {
					for _, val := range col {
						if val == common.NONE_COLOR {
							emptyCount += 1
						}
					}
				}

				// If there are no empty spaces or no one took production, skip it
				if emptyCount == 0 || produceGoodsPlayer == "" {
					err := handler.executeGoodsGrowthPhase(gameMap)
					if err != nil {
						return err
					}
				} else {
					drawCount := 2
					if emptyCount < 2 {
						drawCount = emptyCount
					}

					handler.activePlayer = produceGoodsPlayer
					gameState.ProductionCubes = make([]common.Color, drawCount)

					for n := 0; n < drawCount; n++ {
						var err error
						gameState.ProductionCubes[n], err = gameState.DrawCube(handler.randProvider)
						if err != nil {
							return err
						}

						handler.Log("A %s cube was drawn for the production action.",
							gameState.ProductionCubes[n].String())
					}
				}
			}
		} else {
			handler.activePlayer = nextPlayer
		}
	}

	return nil
}

func (handler *confirmMoveHandler) executeIncomeAndExpenses() error {
	gameState := handler.gameState
	for _, player := range gameState.PlayerOrder {
		cash := gameState.PlayerCash[player] + gameState.PlayerIncome[player] - gameState.PlayerLoco[player] - gameState.PlayerShares[player]
		handler.Log("%s gains $%d in income, pays $%d for loco and $%d for shares.",
			handler.PlayerNick(player), gameState.PlayerIncome[player], gameState.PlayerLoco[player], gameState.PlayerShares[player])
		if cash < 0 {
			gameState.PlayerCash[player] = 0
			gameState.PlayerIncome[player] += cash // cash is negative, so this drops income by the deficit
			handler.Log("%s loses %d in income to pay for excess expenses.", handler.PlayerNick(player), -1*cash)
			// FIXME: Handle bankruptcy
		} else {
			gameState.PlayerCash[player] = cash
		}
	}

	for _, player := range gameState.PlayerOrder {
		income := gameState.PlayerIncome[player]
		reduction, err := handler.gameMap.GetIncomeReduction(gameState, player)
		if err != nil {
			return fmt.Errorf("failed to determine income reduction: %v", err)
		}

		if reduction == 0 {
			handler.Log("%s has %d income and takes no income reduction.", handler.PlayerNick(player), income)
		} else {
			gameState.PlayerIncome[player] -= reduction
			handler.Log("%s has %d income and takes %d income reduction, dropping their income to %d.",
				handler.PlayerNick(player), income, reduction, income-reduction)
		}
	}

	return nil
}

func (handler *confirmMoveHandler) handleProduceGoodsAction(produceGoodsAction *ProduceGoodsAction) error {
	gameState := handler.gameState
	if produceGoodsAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.GOODS_GROWTH_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	if len(gameState.ProductionCubes) != len(produceGoodsAction.Destinations) {
		return &HttpError{"number of destinations must match number of produced goods", http.StatusBadRequest}
	}

	for idx, destination := range produceGoodsAction.Destinations {
		if destination.X < 0 || destination.X >= len(gameState.GoodsGrowth) {
			return &HttpError{fmt.Sprintf("invalid destination column: %d", destination.X), http.StatusBadRequest}
		}
		col := gameState.GoodsGrowth[destination.X]
		if destination.Y < 0 || destination.Y >= len(col) {
			return &HttpError{fmt.Sprintf("invalid destination row: %d", destination.Y), http.StatusBadRequest}
		}
		if col[destination.Y] != common.NONE_COLOR {
			return &HttpError{fmt.Sprintf("goods growth location not empty: (%d,%d)", destination.X, destination.Y), http.StatusBadRequest}
		}
		col[destination.Y] = gameState.ProductionCubes[idx]

		var cityLabel string
		if destination.X < 6 {
			cityLabel = fmt.Sprintf("light city %d", destination.X+1)
		} else if destination.X < 12 {
			cityLabel = fmt.Sprintf("dark city %d", destination.X-5)
		} else {
			cityLabel = fmt.Sprintf("new city %c", 'A'+destination.X-12)
		}

		handler.Log("%s adds a %s cube to %s (spot %d) on the goods growth table using the production action.",
			handler.ActivePlayerNick(), gameState.ProductionCubes[idx].String(), cityLabel, destination.Y+1)
	}

	gameState.ProductionCubes = nil
	err := handler.executeGoodsGrowthPhase(handler.gameMap)
	if err != nil {
		return err
	}

	return nil
}

func getCoordinateForUrb(gameState *common.GameState, urbanNum int) common.Coordinate {
	for _, urb := range gameState.Urbanizations {
		if urb.City == urbanNum {
			return urb.Hex
		}
	}

	return common.Coordinate{X: -1, Y: -1}
}

func (handler *confirmMoveHandler) executeGoodsGrowthPhase(gameMap maps.GameMap) error {
	gameState := handler.gameState
	numPlayers := len(gameState.PlayerOrder)

	// Finds the city on the board for the given column (if present), finds the top cube for the goods growth row (if present), and moves it to the board
	pickCubeFromCol := func(n int) {
		var cord common.Coordinate
		if n < 12 {
			cord = gameMap.GetCityHexForGoodsGrowth(n)
		} else {
			cord = getCoordinateForUrb(gameState, n-12)
		}

		if cord.X < 0 || cord.Y < 0 {
			return
		}

		col := gameState.GoodsGrowth[n]
		pickedColor := common.NONE_COLOR
		for i := 0; i < len(col); i++ {
			color := col[i]
			if color != common.NONE_COLOR {
				pickedColor = color
				col[i] = common.NONE_COLOR
				break
			}
		}

		if pickedColor != common.NONE_COLOR {
			gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
				Color: pickedColor,
				Hex:   cord,
			})
			handler.Log("A %s cube was moved from the goods growth chart to hex (%d,%d).",
				pickedColor.String(), cord.X, cord.Y)
		}
	}

	// Light-side
	for i := 0; i < numPlayers; i++ {
		val, err := handler.randProvider.RandN(6)
		if err != nil {
			return fmt.Errorf("failed to get random number: %v", err)
		}
		handler.Log("A %d was rolled for light cities on goods growth.", val+1)
		pickCubeFromCol(val)
		if 2 <= val && val < 6 {
			pickCubeFromCol(val + 10)
		}
	}

	// Dark-side
	for i := 0; i < numPlayers; i++ {
		val, err := handler.randProvider.RandN(6)
		if err != nil {
			return fmt.Errorf("failed to get random number: %v", err)
		}
		handler.Log("A %d was rolled for dark cities on goods growth.", val+1)
		pickCubeFromCol(val + 6)
		if 0 <= val && val < 4 {
			pickCubeFromCol(val + 16)
		}
	}

	err := gameMap.PostGoodsGrowthHook(gameState, handler.randProvider, handler.Log)
	if err != nil {
		return fmt.Errorf("failed to execute post-goods growth hook: %v", err)
	}

	// Advance to next phase
	gameState.TurnNumber += 1
	gameState.GamePhase = common.SHARES_GAME_PHASE
	handler.activePlayer = gameState.PlayerOrder[0]

	return nil
}
