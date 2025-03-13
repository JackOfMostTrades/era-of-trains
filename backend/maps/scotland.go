package maps

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/auction"
	"github.com/JackOfMostTrades/eot/backend/common"
)

type scotlandMap struct {
	*basicMap
}

type ScotlandAuctionPhase struct {
	auction.StandardAuctionPhase
}

func (p *ScotlandAuctionPhase) PreAuctionHook(handler auction.ConfirmMoveHandler) error {
	gameState := handler.GetGameState()
	if len(gameState.PlayerOrder) == 2 {
		if gameState.PlayerActions[gameState.PlayerOrder[0]] == common.TURN_ORDER_PASS_SPECIAL_ACTION {
			// Player order doesn't change, but skip auction
			handler.Log("Auction is skipped because because of special turn order pass behavior.")
			gameState.AuctionState[gameState.PlayerOrder[1]] = -1
			gameState.AuctionState[gameState.PlayerOrder[0]] = -2
		} else if gameState.PlayerActions[gameState.PlayerOrder[1]] == common.TURN_ORDER_PASS_SPECIAL_ACTION {
			handler.Log("Auction is skipped because because of special turn order pass behavior.")
			gameState.AuctionState[gameState.PlayerOrder[0]] = -1
			gameState.AuctionState[gameState.PlayerOrder[1]] = -2
		}
	}

	err := p.StandardAuctionPhase.PreAuctionHook(handler)
	if err != nil {
		return err
	}
	return nil
}

func (*scotlandMap) GetAuctionPhase() auction.AuctionPhase {
	return &ScotlandAuctionPhase{}
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

func (m *scotlandMap) GetTeleportLink(gameState *common.GameState, src common.Coordinate, direction common.Direction) (*common.Coordinate, common.Direction) {
	dest, destDir := m.basicMap.GetTeleportLink(gameState, src, direction)
	if dest != nil {
		return dest, destDir
	}

	for _, urb := range gameState.Urbanizations {
		if urb.Hex.X == 0 && urb.Hex.Y == 13 {
			if src.X == 0 && src.Y == 13 && direction == common.NORTH_EAST {
				return &common.Coordinate{X: 1, Y: 12}, common.SOUTH_WEST
			}
			if src.X == 1 && src.Y == 12 && direction == common.SOUTH_WEST {
				return &common.Coordinate{X: 0, Y: 13}, common.NORTH_EAST
			}
		}
	}
	return nil, 0
}

func (m *scotlandMap) GetTeleportLinkBuildCost(gameState *common.GameState, player string, hex common.Coordinate, direction common.Direction) int {
	for _, urb := range gameState.Urbanizations {
		if urb.Hex.X == 0 && urb.Hex.Y == 13 {
			if hex.X == 0 && hex.Y == 13 && direction == common.NORTH_EAST {
				return 2
			}
			if hex.X == 1 && hex.Y == 12 && direction == common.SOUTH_WEST {
				return 2
			}
		}
	}
	return m.basicMap.GetTeleportLinkBuildCost(gameState, player, hex, direction)
}

func loadScotlandMap() (GameMap, error) {
	b, err := loadBasicMap("maps/scotland.json")
	if err != nil {
		return nil, err
	}
	return &scotlandMap{b}, nil
}
