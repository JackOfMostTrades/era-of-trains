package main

import (
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type testMap struct {
	maps.AbstractGameMapImpl
	hexes     [][]maps.HexType
	cityColor [][]common.Color
}

func (t *testMap) GetWidth() int {
	return len(t.hexes[0])
}

func (t *testMap) GetHeight() int {
	return len(t.hexes)
}

func (t *testMap) GetHexType(hex common.Coordinate) maps.HexType {
	return t.hexes[hex.Y][hex.X]
}

func (t *testMap) GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color {
	return t.cityColor[hex.Y][hex.X]
}

func (t *testMap) GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate {
	panic("unimplemented")
}

func (t *testMap) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	panic("unimplemented")
}

func TestAttemptTrackPlacement(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 1},
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
			},
			{
				Hex:   common.Coordinate{X: 1, Y: 0},
				Track: [2]common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
			},
			{
				Hex:   common.Coordinate{X: 1, Y: 1},
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestAttemptTrackPlacementEngineer(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]common.SpecialAction{playerId: common.ENGINEER_SPECIAL_ACTION},
	}
	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 1},
				Track: [2]common.Direction{common.NORTH_EAST, common.SOUTH_WEST},
			}, {
				Hex:   common.Coordinate{X: 1, Y: 0},
				Track: [2]common.Direction{common.SOUTH_EAST, common.SOUTH_WEST},
			}, {
				Hex:   common.Coordinate{X: 1, Y: 1},
				Track: [2]common.Direction{common.SOUTH_EAST, common.NORTH_WEST},
			}, {
				Hex:   common.Coordinate{X: 2, Y: 2},
				Track: [2]common.Direction{common.NORTH_EAST, common.NORTH_WEST},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 2, Y: 1}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.NORTH_WEST, common.NORTH_WEST, common.SOUTH_WEST, common.SOUTH_WEST}, link.Steps)
}

func TestTrackFromCityToTown(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 0},
				Track: common.SOUTH_EAST,
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 1},
				Track: [2]common.Direction{common.NORTH_EAST, common.NORTH_WEST},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.NORTH_WEST}, link.Steps)
}

func TestTrackFromTownToCity(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 0},
				Track: common.SOUTH_EAST,
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 1},
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestAdjacentTownAndCity(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Hex:   common.Coordinate{X: 0, Y: 0},
				Track: common.SOUTH_EAST,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST}, link.Steps)
}

func TestUrbanizeAndConnect(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:     common.BUILDING_GAME_PHASE,
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]common.SpecialAction{playerId: common.URBANIZATION_SPECIAL_ACTION},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		Urbanization: &common.Urbanization{
			Hex:  common.Coordinate{X: 0, Y: 0},
			City: 0,
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestBuildThroughTown(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.TOWN_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.SOUTH_WEST,
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
			{
				Track: common.SOUTH_EAST,
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestLolipopToTown(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.TOWN_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.SOUTH_WEST,
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestLolipopFromTown(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.TOWN_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.NORTH_WEST,
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.SOUTH_EAST, common.SOUTH_WEST},
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
			{
				Track: [2]common.Direction{common.NORTH_EAST, common.NORTH_WEST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 1}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.NORTH_WEST, common.SOUTH_WEST, common.NORTH_WEST}, link.Steps)
}

func TestUpgradeToComplex(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 2},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 3}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.NORTH_WEST, common.SOUTH_WEST}, link.Steps)
	assert.Equal(t, 8, gameState.PlayerCash[playerId])

	err = handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 4},
			},
			{
				Track: [2]common.Direction{common.SOUTH, common.NORTH},
				Hex:   common.Coordinate{X: 1, Y: 2},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link = gameState.Links[1]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 3}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.NORTH, common.NORTH}, link.Steps)
	assert.Equal(t, 3, gameState.PlayerCash[playerId])
}

func TestIssue1Regression(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.TOWN_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}

	gameState := &common.GameState{
		PlayerCash: map[string]int{
			playerId: 10,
		},
		GamePhase: common.BUILDING_GAME_PHASE,
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{Track: common.SOUTH_WEST, Hex: common.Coordinate{X: 0, Y: 3}},
			{Track: common.NORTH, Hex: common.Coordinate{X: 0, Y: 3}},
		},
		TrackPlacements: []*TrackPlacement{
			{Track: [2]common.Direction{common.NORTH_EAST, common.SOUTH}, Hex: common.Coordinate{X: 0, Y: 1}},
			{Track: [2]common.Direction{common.NORTH_EAST, common.SOUTH}, Hex: common.Coordinate{X: 0, Y: 4}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 3, gameState.PlayerCash[playerId])
	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 3}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.SOUTH}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.SOUTH}, link.Steps)
}

func TestDirectComplex(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}

	gameState := &common.GameState{
		PlayerCash: map[string]int{
			playerId: 10,
		},
		GamePhase: common.BUILDING_GAME_PHASE,
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{Track: [2]common.Direction{common.SOUTH, common.NORTH_EAST}, Hex: common.Coordinate{X: 0, Y: 3}},
			{Track: [2]common.Direction{common.SOUTH_WEST, common.NORTH}, Hex: common.Coordinate{X: 1, Y: 2}},
			{Track: [2]common.Direction{common.NORTH_EAST, common.SOUTH_EAST}, Hex: common.Coordinate{X: 0, Y: 1}},
			{Track: [2]common.Direction{common.NORTH_WEST, common.SOUTH_EAST}, Hex: common.Coordinate{X: 1, Y: 2}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, gameState.PlayerCash[playerId])
	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 5}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.NORTH, common.NORTH_EAST, common.NORTH}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.SOUTH_EAST, common.SOUTH_EAST}, link.Steps)
}

func TestUrbCompletesLink(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.TOWN_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:     common.BUILDING_GAME_PHASE,
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]common.SpecialAction{playerId: common.URBANIZATION_SPECIAL_ACTION},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)

	err = handler.performBuildAction(&BuildAction{
		Urbanization: &common.Urbanization{
			Hex:  common.Coordinate{X: 1, Y: 0},
			City: 0,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link = gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestExtendIncompleteTrack(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
			{
				Track: [2]common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST}, link.Steps)

	err = handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link = gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestExtendIncompleteUnownedTrack(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
			{
				Track: [2]common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST}, link.Steps)

	err = handler.performBuildAction(&BuildAction{})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link = gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, "", link.Owner)

	err = handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link = gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestRedirectAndCompleteTrack(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.SOUTH},
				Owner:     "",
				Complete:  false,
			},
		},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackRedirects: []*TrackRedirect{
			{
				Track: common.NORTH_EAST,
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestIssue18Regression(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 1, Y: 0},
				Steps:     []common.Direction{common.SOUTH_WEST, common.NORTH_WEST},
				Owner:     playerId,
				Complete:  false,
			},
		},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.SOUTH_EAST,
				Hex:   common.Coordinate{X: 0, Y: 0},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.NORTH_WEST}, link.Steps)
}

func TestIssue26Regression(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.CITY_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST},
				Owner:     playerId,
				Complete:  false,
			},
			{
				SourceHex: common.Coordinate{X: 2, Y: 0},
				Steps:     []common.Direction{common.SOUTH_WEST, common.SOUTH_WEST},
				Owner:     "",
				Complete:  false,
			},
		},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackRedirects: []*TrackRedirect{
			{
				Track: common.NORTH_WEST,
				Hex:   common.Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
}

func TestTownToNowhere(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.SOUTH_EAST,
				Hex:   common.Coordinate{X: 0, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	var invalidMove *invalidMoveError
	require.ErrorAs(t, err, &invalidMove)
	assert.Equal(t, invalidMove.Error(), "all of a player's links must trace back over a player's track to a city")
}

func TestTownToTown(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.TOWN_HEX_TYPE, maps.TOWN_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: common.SOUTH_EAST,
				Hex:   common.Coordinate{X: 0, Y: 0},
			},
			{
				Track: common.SOUTH_WEST,
				Hex:   common.Coordinate{X: 1, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]common.Direction{common.NORTH_WEST, common.NORTH_EAST},
				Hex:   common.Coordinate{X: 0, Y: 1},
			},
		},
	})
	var invalidMove *invalidMoveError
	require.ErrorAs(t, err, &invalidMove)
	assert.Equal(t, invalidMove.Error(), "all of a player's links must trace back over a player's track to a city")
}
