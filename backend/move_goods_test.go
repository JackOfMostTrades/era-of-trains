package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBasicMoveGoods(t *testing.T) {
	testCase := func(expectedError *HttpError, moveAction *MoveGoodsAction) func(t *testing.T) {
		return func(t *testing.T) {
			playerId := "player1"
			playerTwo := "player2"
			theMap := &BasicMap{
				Width:  2,
				Height: 2,
				Hexes: [][]HexType{
					{CITY_HEX_TYPE, CITY_HEX_TYPE},
					{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
				},
				Cities: []BasicCity{
					{
						Color:      PURPLE,
						Coordinate: Coordinate{X: 0, Y: 0},
					},
					{
						Color:      BLUE,
						Coordinate: Coordinate{X: 1, Y: 0},
					},
				},
			}
			gameState := &GameState{
				PlayerOrder: []string{playerId, playerTwo},
				GamePhase:   MOVING_GOODS_GAME_PHASE,
				Links: []*Link{
					{
						SourceHex: Coordinate{X: 0, Y: 0},
						Steps:     []Direction{SOUTH_EAST, NORTH_EAST},
						Complete:  true,
						Owner:     playerId,
					},
				},
				Cubes: []*BoardCube{
					{
						Color: BLUE,
						Hex:   Coordinate{X: 0, Y: 0},
					},
					{
						Color: YELLOW,
						Hex:   Coordinate{X: 0, Y: 0},
					},
				},
				PlayerLoco:   map[string]int{playerId: 1},
				PlayerIncome: map[string]int{playerId: 0},
			}

			handler := &confirmMoveHandler{
				theMap:       theMap,
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
		StartingLocation: Coordinate{X: 0, Y: 0},
		Color:            BLUE,
		Path:             []Direction{SOUTH_EAST},
	}))
	t.Run("bad target color move", testCase(&HttpError{code: http.StatusBadRequest, description: "ending city must match cube color"}, &MoveGoodsAction{
		StartingLocation: Coordinate{X: 0, Y: 0},
		Color:            YELLOW,
		Path:             []Direction{SOUTH_EAST},
	}))
}

func TestCannotMoveThroughMatchingCityColor(t *testing.T) {
	playerId := "player1"
	playerTwo := "player2"
	theMap := &BasicMap{
		Width:  3,
		Height: 2,
		Hexes: [][]HexType{
			{CITY_HEX_TYPE, CITY_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 0},
			},
			{
				Color:      BLUE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
			{
				Color:      BLUE,
				Coordinate: Coordinate{X: 2, Y: 0},
			},
		},
	}
	gameState := &GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   MOVING_GOODS_GAME_PHASE,
		Links: []*Link{
			{
				SourceHex: Coordinate{X: 0, Y: 0},
				Steps:     []Direction{SOUTH_EAST, NORTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: Coordinate{X: 1, Y: 0},
				Steps:     []Direction{SOUTH_EAST, NORTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*BoardCube{
			{
				Color: BLUE,
				Hex:   Coordinate{X: 0, Y: 0},
			},
		},
		PlayerLoco:   map[string]int{playerId: 2},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: Coordinate{X: 0, Y: 0},
		Color:            BLUE,
		Path:             []Direction{SOUTH_EAST, SOUTH_EAST},
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
	theMap := &BasicMap{
		Width:  3,
		Height: 4,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, CITY_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 1},
			},
			{
				Color:      BLUE,
				Coordinate: Coordinate{X: 1, Y: 1},
			},
			{
				Color:      BLUE,
				Coordinate: Coordinate{X: 2, Y: 1},
			},
			{
				Color:      RED,
				Coordinate: Coordinate{X: 0, Y: 3},
			},
		},
	}
	gameState := &GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   MOVING_GOODS_GAME_PHASE,
		Links: []*Link{
			{
				SourceHex: Coordinate{X: 0, Y: 1},
				Steps:     []Direction{NORTH_EAST, SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: Coordinate{X: 1, Y: 1},
				Steps:     []Direction{NORTH_EAST, SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: Coordinate{X: 2, Y: 1},
				Steps:     []Direction{SOUTH_WEST, NORTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: Coordinate{X: 1, Y: 1},
				Steps:     []Direction{SOUTH_WEST, SOUTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*BoardCube{
			{
				Color: RED,
				Hex:   Coordinate{X: 0, Y: 1},
			},
		},
		PlayerLoco:   map[string]int{playerId: 4},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: Coordinate{X: 0, Y: 1},
		Color:            RED,
		Path:             []Direction{NORTH_EAST, NORTH_EAST, SOUTH_WEST, SOUTH_WEST},
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
	theMap := &BasicMap{
		Width:  2,
		Height: 3,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 1},
			},
			{
				Color:      BLUE,
				Coordinate: Coordinate{X: 1, Y: 1},
			},
		},
	}
	gameState := &GameState{
		PlayerOrder: []string{playerId, playerTwo},
		GamePhase:   MOVING_GOODS_GAME_PHASE,
		Links: []*Link{
			{
				SourceHex: Coordinate{X: 0, Y: 1},
				Steps:     []Direction{NORTH_EAST, SOUTH_EAST},
				Complete:  true,
				Owner:     playerId,
			},
			{
				SourceHex: Coordinate{X: 1, Y: 1},
				Steps:     []Direction{SOUTH_WEST, NORTH_WEST},
				Complete:  true,
				Owner:     playerId,
			},
		},
		Cubes: []*BoardCube{
			{
				Color: PURPLE,
				Hex:   Coordinate{X: 0, Y: 1},
			},
		},
		PlayerLoco:   map[string]int{playerId: 2},
		PlayerIncome: map[string]int{playerId: 0},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.handleMoveGoodsAction(&MoveGoodsAction{
		StartingLocation: Coordinate{X: 0, Y: 1},
		Color:            PURPLE,
		Path:             []Direction{NORTH_EAST, SOUTH_WEST},
	})
	if httpErr, ok := err.(*HttpError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.code)
		assert.Equal(t, "cannot repeat a city in the delivery path", httpErr.description)
	} else {
		assert.Fail(t, "Invalid error: %v", err)
	}
}
