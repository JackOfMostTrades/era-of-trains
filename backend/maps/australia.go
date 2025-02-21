package maps

import (
	"fmt"
	"slices"

	"github.com/JackOfMostTrades/eot/backend/common"
)

type australiaMap struct {
	*basicMap
}

func (b *australiaMap) GetTotalBuildCost(gameState *common.GameState, player string, costs []int) int {
	baseCost := b.basicMap.GetTotalBuildCost(gameState, player, costs)

	action := gameState.PlayerActions[player]
	if action != common.ENGINEER_SPECIAL_ACTION {
		return baseCost
	}

	if len(costs) == 0 {
		return baseCost
	}
	return baseCost - slices.Max(costs)
}

func (b *australiaMap) GetDeliveryBonus(coordinate common.Coordinate, color common.Color) int {
	if coordinate.X == 1 && coordinate.Y == 16 {
		return 3
	}
	return 0
}

func (*australiaMap) GetBuildLimit(gameState *common.GameState, player string) (int, error) {
	action := gameState.PlayerActions[player]
	if action == common.URBANIZATION_SPECIAL_ACTION {
		return 2, nil
	}
	return 3, nil
}

func (b *australiaMap) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	count, err := gameState.PullCube(common.BLUE, 12)

	if err != nil {
		return err
	}

	if count != 12 {
		return fmt.Errorf("failed to find 12 blue cubes for Australia, found %d", count)
	}

	for y := 0; y < len(b.Hexes); y++ {
		for x := 0; x < len(b.Hexes[y]); x++ {
			hex := b.Hexes[y][x]
			if hex.HexType != CITY_HEX_TYPE {
				continue
			}
			if len(hex.GoodsGrowth) < 1 || hex.GoodsGrowth[0] < 6 {
				continue
			}

			coord := common.Coordinate{X: x, Y: y}
			gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
				Color: common.BLUE,
				Hex:   coord,
			}, &common.BoardCube{
				Color: common.BLUE,
				Hex:   coord,
			})
		}
	}
	return b.basicMap.PopulateStartingCubes(gameState, randProvider)
}

func loadAustraliaMap() (GameMap, error) {
	b, err := loadBasicMap("maps/australia.json")
	if err != nil {
		return nil, err
	}
	return &australiaMap{b}, nil
}
