package auction

import (
	"errors"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/api"
	"github.com/JackOfMostTrades/eot/backend/common"
	"net/http"
	"slices"
)

type ConfirmMoveHandler interface {
	Log(format string, a ...any)
	GetGameState() *common.GameState
	GetActivePlayer() string
	SetActivePlayer(activePlayer string)
	PlayerNick(playerId string) string
}

type AuctionPhase interface {
	PreAuctionHook(handler ConfirmMoveHandler) error
	HandleBid(handler ConfirmMoveHandler, bidAction *api.BidAction) error
}

type StandardAuctionPhase struct{}

func (s *StandardAuctionPhase) PreAuctionHook(handler ConfirmMoveHandler) error {
	err := s.advanceCurrentPlayerForBidPhase(handler, -1)
	if err != nil {
		return err
	}
	return nil
}

func (s *StandardAuctionPhase) HandleBid(handler ConfirmMoveHandler, bidAction *api.BidAction) error {
	gameState := handler.GetGameState()
	if bidAction == nil {
		return &api.HttpError{"missing bid action", http.StatusBadRequest}
	}
	if gameState.GamePhase != common.AUCTION_GAME_PHASE {
		return &api.HttpError{fmt.Sprintf("invalid action for current phase %d", gameState.GamePhase), http.StatusPreconditionFailed}
	}

	currentPlayer := handler.GetActivePlayer()

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
		cashToPay := calculateCashToPayForBid(lastBid, passCount, len(gameState.PlayerOrder))

		gameState.PlayerCash[currentPlayer] -= cashToPay
		// Set auction state to pass order (-1 first to pass, -2 second to pass, etc)
		gameState.AuctionState[currentPlayer] = (-1 * passCount) - 1

		handler.Log("%s passes, becoming player number %d and paying $%d based on their bid of $%d.", handler.PlayerNick(currentPlayer),
			len(gameState.PlayerOrder)-passCount, cashToPay, lastBid)

	} else if bidAction.Amount == 0 {
		// Bid amount of 0 indicates use of turn-order-pass
		handler.Log("%s uses turn order pass.", handler.PlayerNick(currentPlayer))

		if gameState.PlayerActions[currentPlayer] != common.TURN_ORDER_PASS_SPECIAL_ACTION {
			return &api.HttpError{"current player cannot use turn order pass", http.StatusBadRequest}
		}

		// Do not update this user's bid amount, we just advance the active player
		// Remove user's turn-order-pass action
		gameState.PlayerActions[currentPlayer] = ""

	} else {
		// User is increasing their bid
		playerCash := gameState.PlayerCash[currentPlayer]
		if bidAction.Amount > playerCash {
			return &api.HttpError{fmt.Sprintf("bid amount [%d] greater than player's cash on hand %d", bidAction.Amount, playerCash), http.StatusBadRequest}
		}

		currentHighBid := 0
		for _, bidAmount := range gameState.AuctionState {
			if bidAmount > 0 && bidAmount > currentHighBid {
				currentHighBid = bidAmount
			}
		}
		if bidAction.Amount <= currentHighBid {
			return &api.HttpError{fmt.Sprintf("bid amount [%d] not higher than current high bid %d", bidAction.Amount, currentHighBid), http.StatusBadRequest}
		}

		// Update this user's bid
		gameState.AuctionState[currentPlayer] = bidAction.Amount

		handler.Log("%s bids $%d.", handler.PlayerNick(currentPlayer), bidAction.Amount)
	}

	currentPlayerPos := slices.Index(gameState.PlayerOrder, currentPlayer)
	if currentPlayerPos == -1 {
		return fmt.Errorf("failed to determine current player turn position")
	}
	err := s.advanceCurrentPlayerForBidPhase(handler, currentPlayerPos)
	if err != nil {
		return err
	}

	return nil
}

func (s *StandardAuctionPhase) advanceCurrentPlayerForBidPhase(handler ConfirmMoveHandler, currentPlayerPos int) error {
	gameState := handler.GetGameState()
	if gameState.GamePhase != common.AUCTION_GAME_PHASE {
		return fmt.Errorf("cannot call advanceCurrentPlayerForBidPhase() during this game phase: %d", gameState.GamePhase)
	}

	// How many users have already passed
	passCount := 0
	for _, bidAmount := range gameState.AuctionState {
		if bidAmount < 0 {
			passCount += 1
		}
	}

	currentHighBid := -1
	for _, bidAmount := range gameState.AuctionState {
		if bidAmount > 0 && bidAmount > currentHighBid {
			currentHighBid = bidAmount
		}
	}

	nextPlayer := ""
	for i := 1; i < len(gameState.PlayerOrder); i++ {
		userId := gameState.PlayerOrder[(currentPlayerPos+i)%len(gameState.PlayerOrder)]
		userBid := gameState.AuctionState[userId]
		// Users who have passed or have the current high bid do not go
		if userBid < 0 || userBid == currentHighBid {
			continue
		}
		// If the user does not have enough to outbid and does not have TOP, they auto-pass
		if gameState.PlayerCash[userId] <= 0 || gameState.PlayerCash[userId] <= currentHighBid {
			if gameState.PlayerActions[userId] != common.TURN_ORDER_PASS_SPECIAL_ACTION {
				cashToPay := calculateCashToPayForBid(gameState.AuctionState[userId], passCount, len(gameState.PlayerOrder))
				gameState.PlayerCash[userId] -= cashToPay
				gameState.AuctionState[userId] = (-1 * passCount) - 1
				passCount += 1

				handler.Log("%s automatically passes as they cannot outbid the current bid",
					handler.PlayerNick(userId))

				continue
			}
		}

		nextPlayer = userId
		break
	}

	// If all but one player has passed, that player implicitly passes since there's no one else left
	if passCount == len(gameState.PlayerOrder)-1 {
		// Implicitly pass the remaining player
		for _, playerId := range gameState.PlayerOrder {
			if bidAmount := gameState.AuctionState[playerId]; bidAmount >= 0 {
				gameState.PlayerCash[playerId] -= calculateCashToPayForBid(bidAmount, passCount, len(gameState.PlayerOrder))
				gameState.AuctionState[playerId] = (-1 * passCount) - 1
				passCount += 1

				handler.Log("%s becomes first player as last player to not pass, and pays $%d.",
					handler.PlayerNick(playerId), bidAmount)
			}
		}
	}

	// All players have passed, so advance to next phase
	if passCount == len(gameState.PlayerOrder) {
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
		handler.SetActivePlayer(gameState.PlayerOrder[0])
		// Advance the game phase
		gameState.GamePhase = common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
		// Force-remove any chosen special actions as we advance into that phase
		for userId := range gameState.PlayerActions {
			gameState.PlayerActions[userId] = ""
		}
	} else {
		// Otherwise set the active player to the next player determined above.
		// If nextPlayer is empty, it means we looped already the way back around to currentPlayerPos; this should only
		// happen when using TOP and there are 2 players left. See #13
		if nextPlayer == "" {
			if currentPlayerPos == -1 {
				return errors.New("unable to pick next player to bid, but no one has bid yet")
			}
			nextPlayer = gameState.PlayerOrder[currentPlayerPos]
		}
		handler.SetActivePlayer(nextPlayer)
	}

	return nil
}

func calculateCashToPayForBid(bidAmount int, passCount int, playerCount int) int {
	if passCount == 0 {
		// Last player does not pay
		return 0
	} else if (playerCount - passCount) > 2 {
		// In the middle, pay half price (rounded up)
		return bidAmount/2 + (bidAmount % 2)
	} else {
		// Only two players left to pass, pay full price of the last bid
		return bidAmount
	}
}
