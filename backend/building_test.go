package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAttemptTrackPlacement(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  3,
		Height: 2,
		Hexes: [][]HexType{
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      RED,
				Coordinate: Coordinate{X: 0, Y: 0},
			},
			BasicCity{
				Color:      BLUE,
				Coordinate: Coordinate{X: 2, Y: 0},
			},
		},
	}
	gameState := &GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState, playerId)
	err := performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 0, Y: 1},
		Track: [2]Direction{NORTH_WEST, NORTH_EAST},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 1, Y: 0},
		Track: [2]Direction{SOUTH_WEST, SOUTH_EAST},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 1, Y: 1},
		Track: [2]Direction{NORTH_WEST, NORTH_EAST},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST, SOUTH_EAST, NORTH_EAST}, link.Steps)
}

func TestAttemptTrackPlacementEngineer(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  3,
		Height: 3,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, CITY_HEX_TYPE},
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 2},
			},
			BasicCity{
				Color:      RED,
				Coordinate: Coordinate{X: 2, Y: 1},
			},
		},
	}
	gameState := &GameState{
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]SpecialAction{playerId: ENGINEER_SPECIAL_ACTION},
	}
	performer := newBuildActionPerformer(theMap, gameState, playerId)
	err := performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 0, Y: 1},
		Track: [2]Direction{NORTH_EAST, SOUTH_WEST},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 1, Y: 0},
		Track: [2]Direction{SOUTH_EAST, SOUTH_WEST},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 1, Y: 1},
		Track: [2]Direction{SOUTH_EAST, NORTH_WEST},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 2, Y: 2},
		Track: [2]Direction{NORTH_EAST, NORTH_WEST},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 2, Y: 1}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, NORTH_WEST, NORTH_WEST, SOUTH_WEST, SOUTH_WEST}, link.Steps)
}

func TestTrackFromCityToTown(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{TOWN_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
		},
	}
	gameState := &GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState, playerId)
	err := performer.attemptTownPlacement(&TownPlacement{
		Hex:   Coordinate{X: 0, Y: 0},
		Track: SOUTH_EAST,
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 0, Y: 1},
		Track: [2]Direction{NORTH_EAST, NORTH_WEST},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, NORTH_WEST}, link.Steps)
}

func TestTrackFromTownToCity(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{TOWN_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
		},
	}
	gameState := &GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState, playerId)
	err := performer.attemptTownPlacement(&TownPlacement{
		Hex:   Coordinate{X: 0, Y: 0},
		Track: SOUTH_EAST,
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Hex:   Coordinate{X: 0, Y: 1},
		Track: [2]Direction{NORTH_WEST, NORTH_EAST},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST}, link.Steps)
}

func TestAdjacentTownAndCity(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{TOWN_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 1},
			},
		},
	}
	gameState := &GameState{
		PlayerCash: map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState, playerId)
	err := performer.attemptTownPlacement(&TownPlacement{
		Hex:   Coordinate{X: 0, Y: 0},
		Track: SOUTH_EAST,
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST}, link.Steps)
}

func TestUrbanizeAndConnect(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{TOWN_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
		},
	}
	gameState := &GameState{
		GamePhase:     BUILDING_GAME_PHASE,
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]SpecialAction{playerId: URBANIZATION_SPECIAL_ACTION},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		Urbanization: &Urbanization{
			Hex:  Coordinate{X: 0, Y: 0},
			City: 0,
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{NORTH_WEST, NORTH_EAST},
				Hex:   Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST}, link.Steps)
}

func TestBuildThroughTown(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  3,
		Height: 2,
		Hexes: [][]HexType{
			{CITY_HEX_TYPE, TOWN_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 0},
			},
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 2, Y: 0},
			},
		},
	}
	gameState := &GameState{
		GamePhase:  BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: SOUTH_WEST,
				Hex:   Coordinate{X: 1, Y: 0},
			},
			{
				Track: SOUTH_EAST,
				Hex:   Coordinate{X: 1, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{NORTH_WEST, NORTH_EAST},
				Hex:   Coordinate{X: 0, Y: 1},
			},
			{
				Track: [2]Direction{NORTH_WEST, NORTH_EAST},
				Hex:   Coordinate{X: 1, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST}, link.Steps)
}

func TestLolipopToTown(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{CITY_HEX_TYPE, TOWN_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 0},
			},
		},
	}
	gameState := &GameState{
		GamePhase:  BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: SOUTH_WEST,
				Hex:   Coordinate{X: 1, Y: 0},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{NORTH_WEST, NORTH_EAST},
				Hex:   Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST, NORTH_EAST}, link.Steps)
}

func TestLolipopFromTown(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 2,
		Hexes: [][]HexType{
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, TOWN_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 0},
			},
		},
	}
	gameState := &GameState{
		GamePhase:  BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{
				Track: NORTH_WEST,
				Hex:   Coordinate{X: 1, Y: 1},
			},
		},
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{SOUTH_EAST, SOUTH_WEST},
				Hex:   Coordinate{X: 1, Y: 0},
			},
			{
				Track: [2]Direction{NORTH_EAST, NORTH_WEST},
				Hex:   Coordinate{X: 0, Y: 1},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 1}, link.SourceHex)
	assert.Equal(t, []Direction{NORTH_WEST, SOUTH_WEST, NORTH_WEST}, link.Steps)
}

func TestUpgradeToComplex(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 5,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 3},
			},
		},
	}
	gameState := &GameState{
		GamePhase:  BUILDING_GAME_PHASE,
		PlayerCash: map[string]int{playerId: 10},
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{SOUTH_WEST, SOUTH_EAST},
				Hex:   Coordinate{X: 1, Y: 2},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 3}, link.SourceHex)
	assert.Equal(t, []Direction{NORTH_WEST, SOUTH_WEST}, link.Steps)
	assert.Equal(t, 8, gameState.PlayerCash[playerId])

	err = handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{
				Track: [2]Direction{NORTH, NORTH_EAST},
				Hex:   Coordinate{X: 1, Y: 4},
			},
			{
				Track: [2]Direction{SOUTH, NORTH},
				Hex:   Coordinate{X: 1, Y: 2},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(gameState.Links))
	link = gameState.Links[1]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 3}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, NORTH, NORTH}, link.Steps)
	assert.Equal(t, 3, gameState.PlayerCash[playerId])
}

func TestIssue1Regression(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 7,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{TOWN_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 6},
			},
		},
	}

	gameState := &GameState{
		PlayerCash: map[string]int{
			playerId: 10,
		},
		GamePhase: BUILDING_GAME_PHASE,
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TownPlacements: []*TownPlacement{
			{Track: SOUTH_WEST, Hex: Coordinate{X: 0, Y: 3}},
			{Track: NORTH, Hex: Coordinate{X: 0, Y: 3}},
		},
		TrackPlacements: []*TrackPlacement{
			{Track: [2]Direction{NORTH_EAST, SOUTH}, Hex: Coordinate{X: 0, Y: 1}},
			{Track: [2]Direction{NORTH_EAST, SOUTH}, Hex: Coordinate{X: 0, Y: 4}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 3, gameState.PlayerCash[playerId])
	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 3}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, SOUTH}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, SOUTH}, link.Steps)
}

func TestDirectComplex(t *testing.T) {
	playerId := "player1"
	theMap := &BasicMap{
		Width:  2,
		Height: 6,
		Hexes: [][]HexType{
			{PLAINS_HEX_TYPE, CITY_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{PLAINS_HEX_TYPE, PLAINS_HEX_TYPE},
			{CITY_HEX_TYPE, PLAINS_HEX_TYPE},
		},
		Cities: []BasicCity{
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 1, Y: 0},
			},
			BasicCity{
				Color:      PURPLE,
				Coordinate: Coordinate{X: 0, Y: 5},
			},
		},
	}

	gameState := &GameState{
		PlayerCash: map[string]int{
			playerId: 10,
		},
		GamePhase: BUILDING_GAME_PHASE,
	}

	handler := &confirmMoveHandler{
		theMap:       theMap,
		gameState:    gameState,
		activePlayer: playerId,
	}
	err := handler.performBuildAction(&BuildAction{
		TrackPlacements: []*TrackPlacement{
			{Track: [2]Direction{SOUTH, NORTH_EAST}, Hex: Coordinate{X: 0, Y: 3}},
			{Track: [2]Direction{SOUTH_WEST, NORTH}, Hex: Coordinate{X: 1, Y: 2}},
			{Track: [2]Direction{NORTH_EAST, SOUTH_EAST}, Hex: Coordinate{X: 0, Y: 1}},
			{Track: [2]Direction{NORTH_WEST, SOUTH_EAST}, Hex: Coordinate{X: 1, Y: 2}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 2, gameState.PlayerCash[playerId])
	assert.Equal(t, 2, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 5}, link.SourceHex)
	assert.Equal(t, []Direction{NORTH, NORTH_EAST, NORTH}, link.Steps)
	link = gameState.Links[1]
	assert.Equal(t, false, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 1, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_WEST, SOUTH_EAST, SOUTH_EAST}, link.Steps)
}
