package maps

import (
	"testing"

	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/stretchr/testify/assert"
)

func TestGetTotalBuildCost(t *testing.T) {
	testCase := func(expectedCost int, action common.SpecialAction, costs []int) func(t *testing.T) {
		return func(t *testing.T) {
			playerId := "123"
			gameMap := &australiaMap{&basicMap{}}
			gameState := &common.GameState{
				PlayerActions: map[string]common.SpecialAction{playerId: action},
			}

			cost := gameMap.GetTotalBuildCost(gameState, playerId, costs)
			assert.Equal(t, expectedCost, cost)
		}
	}

	t.Run("regular costs for non-engineer",
		testCase(15, common.FIRST_BUILD_SPECIAL_ACTION, []int{4, 5, 6}))

	t.Run("handles empty arrays gracefully",
		testCase(0, common.ENGINEER_SPECIAL_ACTION, []int{}))

	t.Run("ignores the most expensive element",
		testCase(19, common.ENGINEER_SPECIAL_ACTION, []int{4, 1, 3, 5, 9, 6}))
}

func TestGetBuildLimit(t *testing.T) {
	testCase := func(expectedResult int, action common.SpecialAction) func(t *testing.T) {
		return func(t *testing.T) {
			playerId := "123"
			gameMap := &australiaMap{&basicMap{}}
			gameState := &common.GameState{
				PlayerActions: map[string]common.SpecialAction{playerId: action},
			}

			result, err := gameMap.GetBuildLimit(gameState, playerId)
			assert.Equal(t, expectedResult, result)
			assert.Equal(t, nil, err)
		}
	}

	t.Run("build limitted to 3 as normal", testCase(3, common.FIRST_BUILD_SPECIAL_ACTION))

	t.Run("build limitted to 2 with urbanization", testCase(2, common.URBANIZATION_SPECIAL_ACTION))

	t.Run("engineer does not increase build limits", testCase(3, common.ENGINEER_SPECIAL_ACTION))
}

func TestPopulateStartingCubes(t *testing.T) {
	gameMap := &australiaMap{&basicMap{}}
	gameState := &common.GameState{
		CubeBag: map[common.Color]int{
			common.BLUE:   12,
			common.RED:    12,
			common.PURPLE: 12,
			common.YELLOW: 12,
			common.BLACK:  20,
		},
	}

	err := gameMap.PopulateStartingCubes(gameState, &common.CryptoRandProvider{})
	assert.Equal(t, nil, err)

	for y := 0; y < len(gameMap.Hexes); y++ {
		for x := 0; x < len(gameMap.Hexes[y]); x++ {
			hex := gameMap.Hexes[y][x]
			if hex.HexType != CITY_HEX_TYPE {
				continue
			}
			if len(hex.GoodsGrowth) < 1 || hex.GoodsGrowth[0] < 6 {
				continue
			}

			var cubes []*common.BoardCube
			for _, cube := range gameState.Cubes {
				if cube.Hex.X == x && cube.Hex.Y == y {
					cubes = append(cubes, cube)
				}
			}
			if len(hex.GoodsGrowth) >= 1 && hex.GoodsGrowth[0] >= 6 {
				assert.Equal(t, common.BLUE, cubes[0])
				assert.Equal(t, common.BLUE, cubes[1])
				assert.Equal(t, 4, len(cubes))
			} else {
				assert.Equal(t, 3, len(cubes))
			}
		}
	}
}
