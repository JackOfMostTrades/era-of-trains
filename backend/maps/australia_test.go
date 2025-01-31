package maps

import (
	"testing"

	"github.com/samber/lo"

	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/stretchr/testify/assert"
)

func TestGetTotalBuildCost(t *testing.T) {
	testCase := func(expectedCost int, action common.SpecialAction, townCosts []int, trackCosts []int, teleportCosts []int) func(t *testing.T) {
		return func(t *testing.T) {
			playerId := "123"
			gameMap := &australiaMap{&basicMap{}}
			gameState := &common.GameState{
				PlayerActions: map[string]common.SpecialAction{playerId: action},
			}

			cost := gameMap.GetTotalBuildCost(gameState, playerId, townCosts, trackCosts, teleportCosts)
			assert.Equal(t, expectedCost, cost)
		}
	}

	emptyArray := []int{}

	t.Run("regular costs for non-engineer",
		testCase(15, common.FIRST_BUILD_SPECIAL_ACTION, []int{4}, []int{5}, []int{6}))

	t.Run("handles empty arrays gracefully",
		testCase(0, common.ENGINEER_SPECIAL_ACTION, emptyArray, emptyArray, emptyArray))

	t.Run("ignores the most expensive element",
		testCase(19, common.ENGINEER_SPECIAL_ACTION, []int{4, 1}, []int{3, 5}, []int{9, 6}))
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

	for _, city := range gameMap.Cities {
		cubes := lo.Filter(gameState.Cubes, func(cube *common.BoardCube, index int) bool {
			return cube.Hex == city.Coordinate
		})
		if city.GoodsGrowth[0] >= 6 {
			assert.Equal(t, common.BLUE, cubes[0])
			assert.Equal(t, common.BLUE, cubes[1])
			assert.Equal(t, 4, len(cubes))
		} else {
			assert.Equal(t, 3, len(cubes))
		}
	}
}
