package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/JackOfMostTrades/eot/backend/api"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
)

type confirmMoveHandler struct {
	gameId         string
	gameMap        maps.GameMap
	gameState      *common.GameState
	activePlayer   string
	logs           []string
	reversible     bool
	playerIdToNick map[string]string
	randProvider   common.RandProvider
	gameFinished   bool
}

func (handler *confirmMoveHandler) NumPlayers() int {
	return len(handler.playerIdToNick)
}
func (handler *confirmMoveHandler) GetGameState() *common.GameState {
	return handler.gameState
}
func (handler *confirmMoveHandler) GetActivePlayer() string {
	return handler.activePlayer
}
func (handler *confirmMoveHandler) SetActivePlayer(activePlayer string) {
	handler.activePlayer = activePlayer
}

type invalidMoveError struct {
	description string
}

func (err *invalidMoveError) Error() string {
	return err.description
}

func invalidMoveErr(format string, a ...any) error {
	return &invalidMoveError{fmt.Sprintf(format, a...)}
}

func newConfirmMoveHandler(server *GameServer, gameId string, gameMap maps.GameMap, gameState *common.GameState, activePlayer string) (*confirmMoveHandler, error) {
	handler := &confirmMoveHandler{
		gameId:         gameId,
		gameMap:        gameMap,
		gameState:      gameState,
		activePlayer:   activePlayer,
		playerIdToNick: make(map[string]string),
		randProvider:   server.randProvider,
		gameFinished:   false,
		reversible:     true,
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

func (server *GameServer) confirmMove(ctx *RequestContext, req *api.ConfirmMoveRequest) (resp *api.ConfirmMoveResponse, err error) {
	stmt, err := server.db.Prepare("SELECT owner_user_id,map_name,started,finished,game_state,active_player_id FROM games WHERE id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(req.GameId)
	var ownerUserId string
	var mapName string
	var startedFlag int
	var finishedFlag int
	var gameStateStr string
	var activePlayer string
	err = row.Scan(&ownerUserId, &mapName, &startedFlag, &finishedFlag, &gameStateStr, &activePlayer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &api.HttpError{fmt.Sprintf("invalid game id: %s", req.GameId), http.StatusBadRequest}
		}
		return nil, fmt.Errorf("failed to fetch game row: %v", err)
	}

	if startedFlag == 0 {
		return nil, &api.HttpError{fmt.Sprintf("cannot make a move if game hasn't started yet: %s", req.GameId), http.StatusBadRequest}
	}
	if finishedFlag != 0 {
		return nil, &api.HttpError{fmt.Sprintf("cannot make a move if game has finished: %s", req.GameId), http.StatusBadRequest}
	}

	gameState := new(common.GameState)
	err = json.Unmarshal([]byte(gameStateStr), gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to parse game state: %v", err)
	}
	if activePlayer != ctx.User.Id {
		return nil, &api.HttpError{fmt.Sprintf("user [%s] is not the active player [%s]", ctx.User.Id, activePlayer), http.StatusPreconditionFailed}
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

	if handler.gameFinished {
		finishedFlag = 1
		handler.reversible = false
	}

	// Log the action
	stmt, err = server.db.Prepare("INSERT INTO game_log (game_id,timestamp,user_id,action,description,new_active_player,new_game_state,reversible) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()
	reqString, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialze the request for logging: %v", err)
	}
	_, err = stmt.Exec(req.GameId, time.Now().Unix(), ctx.User.Id, string(reqString), strings.Join(handler.logs, "\n"),
		handler.activePlayer, string(newGameStateStr), boolToInt(handler.reversible))
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

	return &api.ConfirmMoveResponse{}, nil
}

func (handler *confirmMoveHandler) handleAction(req *api.ConfirmMoveRequest) error {
	var err error
	switch req.ActionName {
	case api.SharesActionName:
		err = handler.handleSharesAction(req.SharesAction)
	case api.BidActionName:
		err = handler.handleBidAction(req.BidAction)
	case api.ChooseActionName:
		err = handler.handleChooseAction(req.ChooseAction)
	case api.BuildActionName:
		err = handler.handleBuildAction(req.BuildAction)
	case api.MoveGoodsActionName:
		err = handler.handleMoveGoodsAction(req.MoveGoodsAction)
	case api.ProduceGoodsActionName:
		err = handler.handleProduceGoodsAction(req.ProduceGoodsAction)
	default:
		err = &api.HttpError{fmt.Sprintf("invalid action: %s", req.ActionName), http.StatusBadRequest}
	}
	if err != nil {
		var invalidMove *invalidMoveError
		if errors.As(err, &invalidMove) {
			return &api.HttpError{invalidMove.Error(), http.StatusBadRequest}
		}
		return err
	}
	return nil
}

func (handler *confirmMoveHandler) handleSharesAction(sharesAction *api.SharesAction) error {
	gameState := handler.gameState
	if sharesAction == nil || sharesAction.Amount < 0 {
		return &api.HttpError{"missing shares action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.SHARES_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := handler.activePlayer
	newSharesCount := gameState.PlayerShares[currentPlayer] + sharesAction.Amount
	sharesLimit := handler.gameMap.GetSharesLimit()
	if newSharesCount > sharesLimit {
		return &api.HttpError{fmt.Sprintf("cannot take more than %d shares", sharesLimit), http.StatusBadRequest}
	}
	gameState.PlayerShares[currentPlayer] = newSharesCount
	gameState.PlayerCash[currentPlayer] += 5 * sharesAction.Amount

	handler.Log("%s takes %d shares.", handler.ActivePlayerNick(), sharesAction.Amount)

	currentPlayerPos := slices.Index(gameState.PlayerOrder, currentPlayer)
	if currentPlayerPos == -1 {
		return fmt.Errorf("failed to determine current player turn position")
	}

	err := handler.advanceCurrentPlayerForSharesPhase(currentPlayerPos)
	if err != nil {
		return err
	}

	return nil
}

func (handler *confirmMoveHandler) advanceCurrentPlayerForSharesPhase(currentPlayerPos int) error {
	gameState := handler.gameState
	if gameState.GamePhase != common.SHARES_GAME_PHASE {
		return fmt.Errorf("cannot call getNextPlayerSharesPhase() during this game phase: %d", gameState.GamePhase)
	}

	sharesLimit := handler.gameMap.GetSharesLimit()
	nextPlayerId := ""
	for i := currentPlayerPos + 1; i < len(gameState.PlayerOrder); i++ {
		playerId := gameState.PlayerOrder[i]
		if gameState.PlayerShares[playerId] < sharesLimit {
			nextPlayerId = playerId
			break
		} else {
			handler.Log("Skipping %s because they are at the share limit (%d) and implicitly take 0 shares",
				handler.PlayerNick(playerId), sharesLimit)
		}
	}
	if nextPlayerId == "" {
		// Advance game phase
		gameState.GamePhase = common.AUCTION_GAME_PHASE
		phase := handler.gameMap.GetAuctionPhase()
		err := phase.PreAuctionHook(handler)
		if err != nil {
			return err
		}
	} else {
		handler.activePlayer = nextPlayerId
	}

	return nil
}

func (handler *confirmMoveHandler) handleBidAction(bidAction *api.BidAction) error {
	phase := handler.gameMap.GetAuctionPhase()
	err := phase.HandleBid(handler, bidAction)
	if err != nil {
		return err
	}
	return nil
}

func (handler *confirmMoveHandler) handleChooseAction(chooseAction *api.ChooseAction) error {
	gameState := handler.gameState
	if chooseAction == nil {
		return &api.HttpError{"missing choose action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	isValid := false
	for _, action := range common.ALL_SPECIAL_ACTIONS {
		if chooseAction.Action == action {
			isValid = true
			break
		}
	}
	if !isValid {
		return &api.HttpError{fmt.Sprintf("invalid action: %s", chooseAction.Action), http.StatusBadRequest}
	}

	isChosen := false
	for _, action := range gameState.PlayerActions {
		if action == chooseAction.Action {
			isChosen = true
			break
		}
	}
	if isChosen {
		return &api.HttpError{fmt.Sprintf("action has already been chosen: %s", chooseAction.Action), http.StatusBadRequest}
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
		return &api.HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
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

func (handler *confirmMoveHandler) handleBuildAction(buildAction *api.BuildAction) error {
	gameState := handler.gameState
	if buildAction == nil {
		return &api.HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.BUILDING_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
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
			return &api.HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
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

func (handler *confirmMoveHandler) handleMoveGoodsAction(moveGoodsAction *api.MoveGoodsAction) error {
	gameState := handler.gameState
	gameMap := handler.gameMap
	if moveGoodsAction == nil {
		return &api.HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.MOVING_GOODS_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	if moveGoodsAction.Loco {
		if gameState.PlayerHasDoneLoco[handler.activePlayer] {
			return &api.HttpError{"player has already done loco this phase", http.StatusBadRequest}
		}
		if gameState.PlayerLoco[handler.activePlayer] >= 6 {
			return &api.HttpError{"player's loco is already at max", http.StatusBadRequest}
		}
		gameState.PlayerHasDoneLoco[handler.activePlayer] = true
		gameState.PlayerLoco[handler.activePlayer] += 1
		handler.Log("%s skipped delivering a good and increased their loco to %d",
			handler.ActivePlayerNick(), gameState.PlayerLoco[handler.activePlayer])
	} else if moveGoodsAction.Color != common.NONE_COLOR {

		deliveryGraph := computeDeliveryGraph(gameState, gameMap)

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
			return &api.HttpError{"no such cube", http.StatusBadRequest}
		}

		handler.Log("%s delivered a %s good cube from %s",
			handler.ActivePlayerNick(), moveGoodsAction.Color.String(), renderHexCoordinate(moveGoodsAction.StartingLocation))

		if len(moveGoodsAction.Path) > gameState.PlayerLoco[handler.activePlayer] {
			return &api.HttpError{"cannot move good further than current loca", http.StatusBadRequest}
		}

		loc := moveGoodsAction.StartingLocation
		seenCities := []common.Coordinate{loc}
		for idx, step := range moveGoodsAction.Path {
			if _, ok := deliveryGraph.hexToDirectionToLink[loc]; !ok {
				return &api.HttpError{"invalid path", http.StatusBadRequest}
			}
			if _, ok := deliveryGraph.hexToDirectionToLink[loc][step]; !ok {
				return &api.HttpError{"invalid path", http.StatusBadRequest}
			}

			link := deliveryGraph.hexToDirectionToLink[loc][step]
			loc = link.destination
			if slices.Index(seenCities, loc) != -1 {
				return &api.HttpError{"cannot repeat a city in the delivery path", http.StatusBadRequest}
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
				return &api.HttpError{"invalid path", http.StatusBadRequest}
			}

			locBlocksCube := cityColor == moveGoodsAction.Color || gameMap.LocationBlocksCubePassage(moveGoodsAction.Color, loc)
			if idx != len(moveGoodsAction.Path)-1 && locBlocksCube {
				return &api.HttpError{"cannot pass through city matching the cube color", http.StatusBadRequest}
			}

			locAcceptsCube := cityColor == moveGoodsAction.Color || gameMap.LocationCanAcceptCube(moveGoodsAction.Color, loc)
			if idx == len(moveGoodsAction.Path)-1 && !locAcceptsCube {
				return &api.HttpError{"ending city must match cube color", http.StatusBadRequest}
			}

			if link.player != "" {
				gameState.PlayerIncome[link.player] += 1
				handler.Log("The cube moved to %s giving one income to %s", renderHexCoordinate(loc), handler.PlayerNick(link.player))
			} else {
				handler.Log("The cube moved to %s; no one gets income for it.", renderHexCoordinate(loc))
			}
		}

		handler.Log("The cube finished its movement at %s", renderHexCoordinate(loc))
		// Put the cube back into the bag
		if gameMap.ShouldPutDeliveryInBag(moveGoodsAction.Color) {
			gameState.CubeBag[moveGoodsAction.Color] += 1
		}

		bonus := gameMap.GetDeliveryBonus(loc, moveGoodsAction.Color)
		if bonus > 0 {
			gameState.PlayerIncome[handler.activePlayer] += bonus
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
			return &api.HttpError{"unable to find current player's turn position", http.StatusInternalServerError}
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
				// If the game was ended, return immediately
				if handler.gameFinished {
					return nil
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

					// Random draw makes action non-reversible
					handler.reversible = false
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
		} else {
			gameState.PlayerCash[player] = cash
		}
	}

	for i := 0; i < len(gameState.PlayerOrder); i++ {
		playerId := gameState.PlayerOrder[i]
		if gameState.PlayerIncome[playerId] < 0 {
			handler.Log("%s goes bankrupt and is eliminated from the game.", handler.PlayerNick(playerId))
			for _, link := range gameState.Links {
				if link.Owner == playerId {
					link.Owner = ""
				}
			}

			gameState.PlayerOrder = DeleteFromSliceOrdered(i, gameState.PlayerOrder)
			i -= 1
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

	if len(gameState.PlayerOrder) == 0 {
		handler.Log("Game ends because all players have gone bankrupt.")
		handler.gameFinished = true
	} else if handler.NumPlayers() > 1 && len(gameState.PlayerOrder) <= 1 {
		handler.Log("Game ends because all but one player has gone bankrupt.")
		handler.gameFinished = true
	}

	return nil
}

func (handler *confirmMoveHandler) handleProduceGoodsAction(produceGoodsAction *api.ProduceGoodsAction) error {
	gameState := handler.gameState
	if produceGoodsAction == nil {
		return &api.HttpError{"missing build action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.GOODS_GROWTH_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	if len(gameState.ProductionCubes) != len(produceGoodsAction.Destinations) {
		return &api.HttpError{"number of destinations must match number of produced goods", http.StatusBadRequest}
	}

	for idx, destination := range produceGoodsAction.Destinations {
		if destination.X < 0 || destination.X >= len(gameState.GoodsGrowth) {
			return &api.HttpError{fmt.Sprintf("invalid destination column: %d", destination.X), http.StatusBadRequest}
		}
		col := gameState.GoodsGrowth[destination.X]
		if destination.Y < 0 || destination.Y >= len(col) {
			return &api.HttpError{fmt.Sprintf("invalid destination row: %d", destination.Y), http.StatusBadRequest}
		}
		if col[destination.Y] != common.NONE_COLOR {
			return &api.HttpError{fmt.Sprintf("goods growth location not empty: (%d,%d)", destination.X, destination.Y), http.StatusBadRequest}
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
	numPlayers := handler.NumPlayers()

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
			handler.Log("A %s cube was moved from the goods growth chart to %s.",
				pickedColor.String(), renderHexCoordinate(cord))
		}
	}

	// Light-side
	diceCount := gameMap.GetGoodsGrowthDiceCount(numPlayers)
	for i := 0; i < diceCount; i++ {
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
	for i := 0; i < diceCount; i++ {
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

	// Advance to next turn
	gameState.TurnNumber += 1
	// Goods growth phase rolls dice and makes the action non-reversible
	handler.reversible = false

	// Determine if the game is over
	turnLimit := gameMap.GetTurnLimit(numPlayers)
	if gameState.TurnNumber > turnLimit {
		handler.gameFinished = true
	}

	gameState.GamePhase = common.SHARES_GAME_PHASE
	err = handler.advanceCurrentPlayerForSharesPhase(-1)
	if err != nil {
		return err
	}

	return nil
}
