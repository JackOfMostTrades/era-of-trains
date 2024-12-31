package maps

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
)

type scotlandMap struct {
	*basicMap
}

func (m *scotlandMap) PreAuctionHook(gameState *common.GameState, log LogFun) error {
	if len(gameState.PlayerOrder) != 2 {
		return nil
	}
	if gameState.PlayerActions[gameState.PlayerOrder[0]] == common.TURN_ORDER_PASS_SPECIAL_ACTION {
		// Player order doesn't change, but skip auction
		log("Auction is skipped because because of special turn order pass behavior.")
		gameState.GamePhase = common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
	} else if gameState.PlayerActions[gameState.PlayerOrder[1]] == common.TURN_ORDER_PASS_SPECIAL_ACTION {
		log("Auction is skipped because because of special turn order pass behavior.")
		gameState.PlayerOrder[0], gameState.PlayerOrder[1] = gameState.PlayerOrder[1], gameState.PlayerOrder[0]
		gameState.GamePhase = common.CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
	}
	return nil
}

func (*scotlandMap) GetTurnLimit(playerCount int) int {
	return 8
}

func (*scotlandMap) GetGoodsGrowthDiceCount(playerCount int) int {
	return 4
}

func (m *scotlandMap) PostBuildActionHook(gameState *common.GameState, player string) error {
	for _, teleportLink := range m.TeleportLinks {
		isBuilt := false
		for _, link := range gameState.Links {
			if link.Complete && link.SourceHex == teleportLink.Left.Hex && link.Steps[0] == teleportLink.Left.Direction {
				isBuilt = true
				break
			}
		}
		if isBuilt {
			isLeftUrbanized := false
			isRightUrbanized := false
			for _, urb := range gameState.Urbanizations {
				if urb.Hex == teleportLink.Left.Hex {
					isLeftUrbanized = true
				}
				if urb.Hex == teleportLink.Right.Hex {
					isRightUrbanized = true
				}
			}
			if !isLeftUrbanized || !isRightUrbanized {
				return fmt.Errorf("cannot build ferry link unless both sides are urbanized")
			}
		}
	}

	return nil
}

func loadScotlandMap() (GameMap, error) {
	b, err := loadBasicMap("maps/scotland.json")
	if err != nil {
		return nil, err
	}
	return &scotlandMap{b}, nil
}
