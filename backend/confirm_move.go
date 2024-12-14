package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Action SpecialAction `json:"action"`
}

type BuildAction struct {
	Links        []*Link       `json:"links"`
	Urbanization *Urbanization `json:"urbanization"`
}

type MoveGoodsAction struct {
	StartingLocation Coordinate  `json:"startingLocation"`
	Color            Color       `json:"color"`
	Path             []Direction `json:"path"`
	Loco             bool        `json:"loco"`
}

type ProduceGoodsAction struct {
	// List (corresponding the cubes in the same order as ProductionCubes in the game state) with X,Y coordinates
	// corresponding to which city (X) and which spot (Y) within that city
	Destinations []Coordinate `json:"destinations"`
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

func (server *GameServer) confirmMove(ctx *RequestContext, req *ConfirmMoveRequest) (resp *ConfirmMoveResponse, err error) {
	stmt, err := server.db.Prepare("SELECT (owner_user_id,num_players,map_name,started,game_state) FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var numPlayers int
	var mapName string
	var startedFlag int
	var gameStateStr string
	err = row.Scan(&ownerUserId, &numPlayers, &mapName, &startedFlag, &gameStateStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	gameState := new(GameState)
	err = json.Unmarshal([]byte(gameStateStr), gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to parse game state: %v", err)
	}
	if gameState.ActivePlayer != ctx.User.Id {
		return nil, &HttpError{fmt.Sprintf("user [%s] is not the active player [%s]", ctx.User.Id, gameState.ActivePlayer), http.StatusPreconditionFailed}
	}

	switch req.ActionName {
	case SharesActionName:
		err = gameState.handleSharesAction(req.SharesAction)
	case BidActionName:
		err = gameState.handleBidAction(req.BidAction)
	case ChooseActionName:
		err = gameState.handleChooseAction(req.ChooseAction)
	case BuildActionName:
		err = gameState.handleBuildAction(server.maps[mapName], req.BuildAction)
	case MoveGoodsActionName:
		err = gameState.handleMoveGoodsAction(server.maps[mapName], req.MoveGoodsAction)
	case ProduceGoodsActionName:
		err = gameState.handleProduceGoodsAction(server.maps[mapName], req.ProduceGoodsAction)
	default:
		err = &HttpError{fmt.Sprintf("invalid action: %s", req.ActionName), http.StatusBadRequest}
	}
	if err != nil {
		return nil, err
	}

	newGameStateStr, err := json.Marshal(gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game state: %v", err)
	}

	stmt, err = server.db.Prepare("UPDATE games SET game_state=? WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(string(newGameStateStr), req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	// Notify the next player that it is their turn if the active player changed
	if gameState.ActivePlayer != ctx.User.Id {
		err = server.notifyPlayer(req.GameId, gameState.ActivePlayer)
		if err != nil {
			return nil, fmt.Errorf("failed to notify user it's their turn: %v", err)
		}
	}

	return &ConfirmMoveResponse{}, nil
}

func (gameState *GameState) handleSharesAction(sharesAction *SharesAction) error {
	if sharesAction == nil || sharesAction.Amount < 0 {
		return &HttpError{"missing shares action", http.StatusBadRequest}
	}
	if gameState.GamePhase != SHARES_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := gameState.ActivePlayer
	newSharesCount := gameState.PlayerShares[currentPlayer] + sharesAction.Amount
	if newSharesCount > 15 {
		return &HttpError{"cannot take more than 15 shares", http.StatusBadRequest}
	}
	gameState.PlayerShares[currentPlayer] = newSharesCount

	nextPlayerId := ""
	for i := 0; i < len(gameState.PlayerOrder)-1; i++ {
		if gameState.PlayerOrder[i] == currentPlayer {
			nextPlayerId = gameState.PlayerOrder[i+1]
			break
		}
	}
	if nextPlayerId == "" {
		// Advance game phase
		gameState.GamePhase = AUCTION_GAME_PHASE
		gameState.ActivePlayer = gameState.PlayerOrder[0]
	} else {
		gameState.ActivePlayer = nextPlayerId
	}

	return nil
}

func (gameState *GameState) handleBidAction(bidAction *BidAction) error {
	if bidAction == nil {
		return &HttpError{"missing bid action", http.StatusBadRequest}
	}
	if gameState.GamePhase != AUCTION_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := gameState.ActivePlayer

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
		var cashToPay int
		if passCount == 0 {
			// Last player does not pay
			cashToPay = 0
		} else if (len(gameState.AuctionState) - passCount) >= 2 {
			// Only two players left to pass, pay full price of the last bid
			cashToPay = gameState.AuctionState[currentPlayer]
		} else {
			// In the middle, pay half price (rounded up)
			bid := gameState.AuctionState[currentPlayer]
			cashToPay = bid/2 + (bid % 2)
		}

		gameState.PlayerCash[currentPlayer] -= cashToPay
		// Set auction state to pass order (-1 first to pass, -2 second to pass, etc)
		gameState.AuctionState[currentPlayer] = (-1 * passCount) - 1

		if passCount == len(gameState.AuctionState)-1 {
			gotoNextPhase = true
		}

	} else if bidAction.Amount == 0 {
		// Bid amount of 0 indicates use of turn-order-pass

		if gameState.PlayerActions[currentPlayer] != TURN_ORDER_PASS_SPECIAL_ACTION {
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
	}

	if gotoNextPhase {
		// Get the new player order from the auction state
		for userId, bidAmount := range gameState.AuctionState {
			gameState.PlayerOrder[(-1*bidAmount)-1] = userId
		}
		// Then reset the auction state
		for _, userId := range gameState.PlayerOrder {
			gameState.AuctionState[userId] = 0
		}
		// Set the active player to the new first player
		gameState.ActivePlayer = gameState.PlayerOrder[0]
		// Advance the game phase
		gameState.GamePhase = CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
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
		gameState.ActivePlayer = nextPlayer
	}

	return nil
}

func (gameState *GameState) handleChooseAction(chooseAction *ChooseAction) error {
	if chooseAction == nil {
		return &HttpError{"missing choose action", http.StatusBadRequest}
	}
	if gameState.GamePhase != CHOOSE_SPECIAL_ACTIONS_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	isValid := false
	for _, action := range []SpecialAction{FIRST_MOVE_SPECIAL_ACTION, FIRST_BUILD_SPECIAL_ACTION, ENGINEER_SPECIAL_ACTION, LOCO_SPECIAL_ACTION, URBANIZATION_SPECIAL_ACTION, PRODUCTION_SPECIAL_ACTION, TURN_ORDER_PASS_SPECIAL_ACTION} {
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
	gameState.PlayerActions[gameState.ActivePlayer] = chooseAction.Action
	// Apply any immediate effects
	if chooseAction.Action == LOCO_SPECIAL_ACTION {
		if gameState.PlayerLoco[gameState.ActivePlayer] < 6 {
			gameState.PlayerLoco[gameState.ActivePlayer] += 1
		}
	}

	// Advance current player
	currentPlayerPosition := -1
	for idx, userId := range gameState.PlayerOrder {
		if userId == gameState.ActivePlayer {
			currentPlayerPosition = idx
			break
		}
	}
	if currentPlayerPosition == -1 {
		return &HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
	}
	// Advance the game phase if this was the last player
	if currentPlayerPosition == len(gameState.PlayerOrder)-1 {
		gameState.GamePhase = BUILDING_GAME_PHASE
		// If anyone has first build, they become the active player
		firstBuildUser := ""
		for userId, action := range gameState.PlayerActions {
			if action == FIRST_BUILD_SPECIAL_ACTION {
				firstBuildUser = userId
				break
			}
		}

		if firstBuildUser == "" {
			gameState.ActivePlayer = gameState.PlayerOrder[0]
		} else {
			gameState.ActivePlayer = firstBuildUser
		}
	} else {
		// Otherwise just advance the active player
		gameState.ActivePlayer = gameState.PlayerOrder[currentPlayerPosition+1]
	}

	return nil
}

func (gameState *GameState) handleBuildAction(theMap *BasicMap, buildAction *BuildAction) error {
	if buildAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != BUILDING_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}
	if buildAction.Urbanization != nil && gameState.PlayerActions[gameState.ActivePlayer] != URBANIZATION_SPECIAL_ACTION {
		return &HttpError{"cannot urbanize without special action", http.StatusBadRequest}
	}

	if buildAction.Urbanization != nil {
		if buildAction.Urbanization.City < 0 || buildAction.Urbanization.City >= 8 {
			return &HttpError{fmt.Sprintf("invalid city: %d", buildAction.Urbanization.City), http.StatusBadRequest}
		}

		for _, existingUrb := range gameState.Urbanizations {
			if existingUrb.Hex == buildAction.Urbanization.Hex {
				return &HttpError{"cannot urbanize on top of existing urbanization", http.StatusBadRequest}
			}
			if existingUrb.City == buildAction.Urbanization.City {
				return &HttpError{"requested city has already been urbanized", http.StatusBadRequest}
			}
		}
		if theMap.Hexes[buildAction.Urbanization.Hex.Y][buildAction.Urbanization.Hex.X] != TOWN_HEX_TYPE {
			return &HttpError{"must urbanize on town hex", http.StatusBadRequest}
		}

		gameState.Urbanizations = append(gameState.Urbanizations, buildAction.Urbanization)
	}

	// FIXME: Handle validating and then adding link and taking user's cash

	// If this was the first build player, just advance to the first player in normal order
	if gameState.PlayerActions[gameState.ActivePlayer] == FIRST_BUILD_SPECIAL_ACTION {
		// If this was the first player anyway, go to next player
		if gameState.PlayerOrder[0] == gameState.ActivePlayer {
			gameState.ActivePlayer = gameState.PlayerOrder[1]
		} else {
			gameState.ActivePlayer = gameState.PlayerOrder[0]
		}
	} else {
		// Otherwise just advance the player, skipping over a player if they have first build
		currentPlayerPosition := -1
		for idx, userId := range gameState.PlayerOrder {
			if userId == gameState.ActivePlayer {
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
			if gameState.PlayerActions[player] == FIRST_BUILD_SPECIAL_ACTION {
				continue
			}
			nextPlayer = player
			break
		}

		// End of phase
		if nextPlayer == "" {
			gameState.GamePhase = MOVING_GOODS_GAME_PHASE
			firstMovePlayer := ""
			for userId, action := range gameState.PlayerActions {
				if action == FIRST_MOVE_SPECIAL_ACTION {
					firstMovePlayer = userId
					break
				}
			}

			if firstMovePlayer == "" {
				gameState.ActivePlayer = firstMovePlayer
			} else {
				gameState.ActivePlayer = gameState.PlayerOrder[0]
			}
		} else {
			gameState.ActivePlayer = nextPlayer
		}
	}

	return nil
}

func (gameState *GameState) handleMoveGoodsAction(theMap *BasicMap, moveGoodsAction *MoveGoodsAction) error {
	if moveGoodsAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != MOVING_GOODS_GAME_PHASE {
		return &HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	if moveGoodsAction.Loco {
		if gameState.PlayerHasDoneLoco[gameState.ActivePlayer] {
			return &HttpError{"player has already done loco this phase", http.StatusBadRequest}
		}
		gameState.PlayerHasDoneLoco[gameState.ActivePlayer] = true
		if gameState.PlayerLoco[gameState.ActivePlayer] < 6 {
			gameState.PlayerLoco[gameState.ActivePlayer] += 1
		}
	} else if moveGoodsAction.Color != NONE_COLOR {

		// FIXME: Handle this delivery

	} else {
		// Pass action, do nothing here
	}

	// If this was the first move player, just advance to the first player in normal order
	if gameState.PlayerActions[gameState.ActivePlayer] == FIRST_MOVE_SPECIAL_ACTION {
		// If this was the first player anyway, go to next player
		if gameState.PlayerOrder[0] == gameState.ActivePlayer {
			gameState.ActivePlayer = gameState.PlayerOrder[1]
		} else {
			gameState.ActivePlayer = gameState.PlayerOrder[0]
		}
	} else {
		// Otherwise just advance the player, skipping over a player if they have first move
		currentPlayerPosition := -1
		for idx, userId := range gameState.PlayerOrder {
			if userId == gameState.ActivePlayer {
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
			if gameState.PlayerActions[player] == FIRST_MOVE_SPECIAL_ACTION {
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
					if action == FIRST_MOVE_SPECIAL_ACTION {
						firstMovePlayer = userId
						break
					}
				}

				if firstMovePlayer == "" {
					gameState.ActivePlayer = gameState.PlayerOrder[0]
				} else {
					gameState.ActivePlayer = firstMovePlayer
				}
			} else {
				// End of phase
				gameState.GamePhase = GOODS_GROWTH_GAME_PHASE
				produceGoodsPlayer := ""
				for userId, action := range gameState.PlayerActions {
					if action == PRODUCTION_SPECIAL_ACTION {
						produceGoodsPlayer = userId
						break
					}
				}

				if produceGoodsPlayer == "" {
					err := gameState.executeGoodsGrowthPhase(theMap)
					if err != nil {
						return err
					}
				} else {
					gameState.ActivePlayer = produceGoodsPlayer
					gameState.ProductionCubes = make([]Color, 2)

					var err error
					gameState.ProductionCubes[0], err = gameState.drawCube()
					if err != nil {
						return err
					}
					gameState.ProductionCubes[1], err = gameState.drawCube()
					if err != nil {
						return err
					}
				}
			}
		} else {
			gameState.ActivePlayer = nextPlayer
		}
	}

	return nil
}

func (gameState *GameState) handleProduceGoodsAction(theMap *BasicMap, produceGoodsAction *ProduceGoodsAction) error {
	if produceGoodsAction == nil {
		return &HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != GOODS_GROWTH_GAME_PHASE {
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
		if col[destination.Y] != NONE_COLOR {
			return &HttpError{fmt.Sprintf("goods growth location not empty: (%d,%d)", destination.X, destination.Y), http.StatusBadRequest}
		}
		col[destination.Y] = gameState.ProductionCubes[idx]
	}

	gameState.ProductionCubes = nil
	err := gameState.executeGoodsGrowthPhase(theMap)
	if err != nil {
		return err
	}

	return nil
}

func (gameState *GameState) getCoordinateForCity(theMap *BasicMap, cityNum int) Coordinate {
	// Basic cities
	if cityNum < 12 {
		for _, city := range theMap.Cities {
			for _, val := range city.GoodsGrowth {
				if val == cityNum {
					return city.Coordinate
				}
			}
		}
	} else {
		// Urbanizations
		urbanNum := cityNum - 12
		for _, urb := range gameState.Urbanizations {
			if urb.City == urbanNum {
				return urb.Hex
			}
		}
	}

	return Coordinate{X: -1, Y: -1}
}

func (gameState *GameState) executeGoodsGrowthPhase(theMap *BasicMap) error {
	numPlayers := len(gameState.PlayerOrder)

	// Finds the city on the board for the given column (if present), finds the top cube for the goods growth row (if present), and moves it to the board
	pickCubeFromCol := func(n int) {
		cord := gameState.getCoordinateForCity(theMap, n)
		if cord.X < 0 || cord.Y < 0 {
			return
		}

		col := gameState.GoodsGrowth[n]
		pickedColor := NONE_COLOR
		for i := 0; i < len(col); i++ {
			color := col[i]
			if color != NONE_COLOR {
				pickedColor = color
				col[i] = NONE_COLOR
				break
			}
		}

		if pickedColor != NONE_COLOR {
			gameState.Cubes = append(gameState.Cubes, &BoardCube{
				Color: pickedColor,
				Hex:   cord,
			})
		}
	}

	// Light-side
	for i := 0; i < numPlayers; i++ {
		val, err := RandN(6)
		if err != nil {
			return fmt.Errorf("failed to get random number: %v", err)
		}
		pickCubeFromCol(val)
		if 2 <= val && val < 6 {
			pickCubeFromCol(val + 10)
		}
	}

	// Dark-side
	for i := 0; i < numPlayers; i++ {
		val, err := RandN(6)
		if err != nil {
			return fmt.Errorf("failed to get random number: %v", err)
		}
		pickCubeFromCol(val + 6)
		if 0 <= val && val < 4 {
			pickCubeFromCol(val + 16)
		}
	}

	// Advance to next phase
	gameState.TurnNumber += 1
	gameState.GamePhase = SHARES_GAME_PHASE
	gameState.ActivePlayer = gameState.PlayerOrder[0]

	// FIXME: Handle end of game

	return nil
}
