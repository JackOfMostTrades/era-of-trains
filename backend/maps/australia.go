package maps

import (
	"fmt"
	"slices"

	"github.com/JackOfMostTrades/eot/backend/common"
)

type australiaMap struct {
	*basicMap
}


func (b *australiaMap) GetTotalBuildCost(gameState *common.GameState, player string,
	redirectCosts []int, townCosts []int, trackCosts []int, teleportCosts []int) int {

	baseCost := b.basicMap.GetTotalBuildCost(
		gameState, player, redirectCosts, townCosts, trackCosts, teleportCosts)

	action := gameState.PlayerActions[player]
	if action != common.ENGINEER_SPECIAL_ACTION {
		return baseCost
	}

	allCosts := []int{}
	allCosts = append(allCosts, redirectCosts...)
	allCosts = append(allCosts, townCosts...)
	allCosts = append(allCosts, trackCosts...)
	allCosts = append(allCosts, teleportCosts...)
	if len(allCosts) == 0 {
		return baseCost
	}
	return baseCost - slices.Max(allCosts)
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
		return fmt.Errorf("failed to find 12 blue cubes for Australia")
	}

	for _, city := range b.Cities {
		if city.GoodsGrowth[0] < 6 {
			continue
		}

		gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
			Color: common.BLUE,
			Hex:   city.Coordinate,
		}, &common.BoardCube{
			Color: common.BLUE,
			Hex:   city.Coordinate,
		})
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