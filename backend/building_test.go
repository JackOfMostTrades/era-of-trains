package main

import (
	"github.com/JackOfMostTrades/eot/backend/api"
	"testing"

	"github.com/JackOfMostTrades/eot/backend/tiles"

	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.STRAIGHT_TRACK_TILE,
					Rotation: 1,
				},
			}, {
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
			}, {
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.STRAIGHT_TRACK_TILE,
					Rotation: 2,
				},
			}, {
				Hex: common.Coordinate{X: 2, Y: 2},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 2}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.NORTH_EAST, common.NORTH_EAST, common.SOUTH_EAST, common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
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
	urb := 0
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex:          common.Coordinate{X: 0, Y: 0},
				Urbanization: &urb,
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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

func TestUrbanizeDangler(t *testing.T) {
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
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 1, Y: 0},
				Steps:     []common.Direction{common.SOUTH_WEST, common.NORTH_WEST},
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
	urb := 0
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex:          common.Coordinate{X: 0, Y: 0},
				Urbanization: &urb,
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_WEST, common.SOUTH_EAST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_WEST, common.NORTH_WEST}, link.Steps)
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_WEST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.NORTH_WEST},
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 2},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
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

	err = handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 4},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.SHARP_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 2},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.BOW_AND_ARROW_TRACK_TILE,
					Rotation: 1,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 3},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_WEST, common.NORTH},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 0,
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 4},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 0,
				},
			},
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
	assert.Equal(t, common.Coordinate{X: 0, Y: 3}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.NORTH, common.NORTH_EAST}, link.Steps)
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 3},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 0,
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.SHARP_CURVE_TRACK_TILE,
					Rotation: 5,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 2},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.BOW_AND_ARROW_TRACK_TILE,
					Rotation: 3,
				},
			},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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

	urb := 0
	err = handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex:          common.Coordinate{X: 1, Y: 0},
				Urbanization: &urb,
			},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
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

	err = handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
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

	err = handler.performBuildAction(&api.BuildAction{})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link = gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, "", link.Owner)

	err = handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
				Steps:     []common.Direction{common.SOUTH_EAST, common.SOUTH_WEST},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 1,
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_EAST},
				},
			},
			{
				Hex: common.Coordinate{X: 1, Y: 0},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{common.SOUTH_WEST},
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
		},
	})
	var invalidMove *invalidMoveError
	require.ErrorAs(t, err, &invalidMove)
	assert.Equal(t, invalidMove.Error(), "all of a player's links must trace back over a player's track to a city")
}

func TestTownTrackLimit(t *testing.T) {
	playerId := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.TOWN_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{
					X: 0,
					Y: 5,
				},
				Steps: []common.Direction{
					common.NORTH,
					common.NORTH_EAST,
					common.NORTH_WEST,
				},
				Complete: true,
				Owner:    playerId,
			},
		},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TownPlacement: &api.TownPlacement{
					Track: []common.Direction{
						common.NORTH,
						common.NORTH_EAST,
						common.NORTH_WEST,
						common.SOUTH_EAST,
						common.SOUTH_WEST},
				},
			},
		},
	})
	var invalidMove *invalidMoveError
	require.ErrorAs(t, err, &invalidMove)
	assert.Equal(t, "cannot build more than four tracks on a town hex", invalidMove.Error())
}

func TestRedirectJoinUnownedLinks(t *testing.T) {
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
				Owner:     "",
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
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 1, Y: 1},

				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.GENTLE_CURVE_TRACK_TILE,
					Rotation: 4,
				},
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

func TestUrbStartOfDangler(t *testing.T) {
	player1 := "player1"
	player2 := "player2"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.TOWN_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.CITY_HEX_TYPE, maps.TOWN_HEX_TYPE, maps.TOWN_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:     common.BUILDING_GAME_PHASE,
		PlayerCash:    map[string]int{player1: 10},
		PlayerActions: map[string]common.SpecialAction{player2: common.URBANIZATION_SPECIAL_ACTION},
		Links: []*common.Link{
			{
				SourceHex: common.Coordinate{X: 0, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.NORTH_EAST},
				Owner:     player1,
				Complete:  true,
			},
			{
				SourceHex: common.Coordinate{X: 1, Y: 0},
				Steps:     []common.Direction{common.SOUTH_EAST, common.SOUTH_EAST},
				Owner:     player1,
				Complete:  false,
			},
		},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: player2,
	}
	urb := 0
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex:          common.Coordinate{X: 1, Y: 0},
				Urbanization: &urb,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, player1, link.Owner)
	assert.Equal(t, common.Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.NORTH_EAST}, link.Steps)

	link = gameState.Links[1]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, player1, link.Owner)
	assert.Equal(t, common.Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []common.Direction{common.SOUTH_EAST, common.SOUTH_EAST}, link.Steps)
}

func TestNoLoopingConnection(t *testing.T) {
	player1 := "player1"
	gameMap := &testMap{
		hexes: [][]maps.HexType{
			{maps.CITY_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
			{maps.PLAINS_HEX_TYPE, maps.PLAINS_HEX_TYPE},
		},
	}
	gameState := &common.GameState{
		GamePhase:  common.BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{player1: 10},
	}

	handler := &confirmMoveHandler{
		gameMap:      gameMap,
		gameState:    gameState,
		activePlayer: player1,
	}
	err := handler.performBuildAction(&api.BuildAction{
		Steps: []*api.BuildStep{
			{
				Hex: common.Coordinate{X: 0, Y: 1},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.SHARP_CURVE_TRACK_TILE,
					Rotation: 2,
				},
			},
			{
				Hex: common.Coordinate{X: 0, Y: 2},
				TrackPlacement: &api.TrackPlacement{
					Tile:     tiles.SHARP_CURVE_TRACK_TILE,
					Rotation: 4,
				},
			},
		},
	})
	var invalidMove *invalidMoveError
	require.ErrorAs(t, err, &invalidMove)
	assert.Equal(t, invalidMove.Error(), "individual links are not allowed to start and end at the same town/city")
}
