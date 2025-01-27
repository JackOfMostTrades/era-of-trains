package main

import (
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBasicMoveGoods(t *testing.T) {
	testCase := func(expectedError *HttpError, moveAction *MoveGoodsAction) func(t *testing.T) {
		return func(t *testing.T) {
			playerId := "player1"
			playerTwo := "player2"
			gameMap := &testMap{
				hexes: [][]maps.HexType{
					{maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE},
					{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
				},
				cityColor: [][]common.Color{
					{common.PURPLE, common.BLUE},
					{common.NONE_COLOR, common.NONE_COLOR},
				},
			}
			gameState := &common.GameState{
				PlayerOrder: []string{playerId, playerTwo},
				GamePhase:   common.MOVING_GOODS_GAME_PHASE,
				CubeBag:     make(map[common.Color]int),
				Links: []*common.Link{
					{
						SourceHex: common.Coordinate{X: 0, Y: 0},
						Steps:     []common.Direction{common.SOUTH_EAST, common.NORTH_EAST},
						Complete:  true,
						Owner:     playerId,
					},
				},
				Cubes: []*common.BoardCube{
					{
						Color: common.BLUE,
						Hex:   common.Coordinate{X: 0, Y: 0},
					},
					{
						Color: common.YELLOW,
						Hex:   common.Coordinate{X: 0, Y: 0},
					},
				},
				PlayerLoco:   map[string]int{playerId: 1},
				PlayerIncome: map[string]int{playerId: 0},
			}

			handler := &confirmMoveHandler{
				gameMap:      gameMap,
				gameState:    gameState,
				activePlayer: playerId,
			}
			err := handler.handleMoveGoodsAction(moveAction)
			if expectedError != nil {
				if httpErr, ok := err.(*HttpError); ok {
					assert.Equal(t, expectedError.code, httpErr.code)
					assert.Equal(t, expectedError.description, httpErr.description)
				} else {
					assert.Fail(t, "Invalid error: %v", err)
				}
			} else {
				assert.Equal(t, 1, gameState.PlayerIncome[playerId])
				assert.Equal(t, 1, len(gameState.Cubes))
			}
		}
	}

	t.Run("simple move", testCase(nil, &MoveGoodsAction{
		StartingLocation: common.Coordinate{X: 0, Y: 0},
		Color:            common.BLUE,
		Path:             []common.Direction{common.SOUTH_EAST},
	}))
	t.Run("bad target color move", testCase(&HttpError{code: http.StatusBadRequest, description: "ending city must match cube color"}, &MoveGoodsAction{
		StartingLocation: common.Coordinate{X: 0, Y: 0},
		Color:            common.YELLOW,
		Path:             []common.Direction{common.SOUTH_EAST},
	}))
}

func TestCannotMoveThroughMatchingCityColor(t *testing.T) {
	playerId := "player1"
	playerTwo := "player2"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
		cityColor: [][]common.Color{
			{common.PURPLE, common.BLUE, common.BLUE},
			{common.NONE_COLOR, common.NONE_COLOR, common.NONE_COLOR},
		},
	}
	gameState := &common.GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   common.MOVING_GOODS_GAME_PHASE,
		CubeBag:     make(map[common.Color]int),
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.NORTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: common.Coordinate{X: 1, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.NORTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*common.BoardCube{
			{
				Color: common.BLUE,
				Hex:   common.Coordinate{X: 0, Y: 0},
			},
		},
		PlayerLoco:   map[string]int{playerId: 2},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: common.Coordinate{X: 0, Y: 0},
		Color:            common.BLUE,
		Path:             []common.Direction{common.SOUTH_EAST, common.SOUTH_EAST},
	})
	if httpErr, ok := err.(*HttpError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.code)
		assert.Equal(t, "cannot pass through city matching the cube color", httpErr.description)
	} else {
		assert.Fail(t, "Invalid error: %v", err)
	}
}

func TestCannotRepeatCity(t *testing.T) {
	playerId := "player1"
	playerTwo := "player2"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
		cityColor: [][]common.Color{
			{common.NONE_COLOR, common.NONE_COLOR, common.NONE_COLOR},
			{common.PURPLE, common.BLUE, common.BLUE},
			{common.NONE_COLOR, common.NONE_COLOR, common.NONE_COLOR},
			{common.RED, common.NONE_COLOR, common.NONE_COLOR},
		},
	}
	gameState := &common.GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   common.MOVING_GOODS_GAME_PHASE,
		CubeBag:     make(map[common.Color]int),
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 1},
				Steps:     []common.Direction{common.NORTH_EAST, common.SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: common.Coordinate{X: 1, Y: 1},
				Steps:     []common.Direction{common.NORTH_EAST, common.SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: common.Coordinate{X: 2, Y: 1},
				Steps:     []common.Direction{common.SOUTH_WEST, common.NORTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: common.Coordinate{X: 1, Y: 1},
				Steps:     []common.Direction{common.SOUTH_WEST, common.SOUTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*common.BoardCube{
			{
				Color: common.RED,
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
		PlayerLoco:   map[string]int{playerId: 4},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: common.Coordinate{X: 0, Y: 1},
		Color:            common.RED,
		Path:             []common.Direction{common.NORTH_EAST, common.NORTH_EAST, common.SOUTH_WEST, common.SOUTH_WEST},
	})
	if httpErr, ok := err.(*HttpError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.code)
		assert.Equal(t, "cannot repeat a city in the delivery path", httpErr.description)
	} else {
		assert.Fail(t, "Invalid error: %v", err)
	}
}

func TestCannotEndInStartingCity(t *testing.T) {
	playerId := "player1"
	playerTwo := "player2"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		}, cityColor: [][]common.Color{
			{common.NONE_COLOR, common.NONE_COLOR},
			{common.PURPLE, common.BLUE},
			{common.NONE_COLOR, common.NONE_COLOR},
		},
	}
	gameState := &common.GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   common.MOVING_GOODS_GAME_PHASE,
		CubeBag:     make(map[common.Color]int),
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 1},
				Steps:     []common.Direction{common.NORTH_EAST, common.SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: common.Coordinate{X: 1, Y: 1},
				Steps:     []common.Direction{common.SOUTH_WEST, common.NORTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*common.BoardCube{
			{
				Color: common.PURPLE,
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
		PlayerLoco:   map[string]int{playerId: 2},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: common.Coordinate{X: 0, Y: 1},
		Color:            common.PURPLE,
		Path:             []common.Direction{common.NORTH_EAST, common.SOUTH_WEST},
	})
	if httpErr, ok := err.(*HttpError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.code)
		assert.Equal(t, "cannot repeat a city in the delivery path", httpErr.description)
	} else {
		assert.Fail(t, "Invalid error: %v", err)
	}
}
