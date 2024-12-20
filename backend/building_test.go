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
		ActivePlayer: playerId,
		PlayerCash:   map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState)
	err := performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_WEST, NORTH_EAST}},
		Hex:    Coordinate{X: 0, Y: 1},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{SOUTH_WEST, SOUTH_EAST}},
		Hex:    Coordinate{X: 1, Y: 0},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_WEST, NORTH_EAST}},
		Hex:    Coordinate{X: 1, Y: 1},
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
		ActivePlayer:  playerId,
		PlayerCash:    map[string]int{playerId: 10},
		PlayerActions: map[string]SpecialAction{playerId: ENGINEER_SPECIAL_ACTION},
	}
	performer := newBuildActionPerformer(theMap, gameState)
	err := performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_EAST, SOUTH_WEST}},
		Hex:    Coordinate{X: 0, Y: 1},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{SOUTH_EAST, SOUTH_WEST}},
		Hex:    Coordinate{X: 1, Y: 0},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{SOUTH_EAST, NORTH_WEST}},
		Hex:    Coordinate{X: 1, Y: 1},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_EAST, NORTH_WEST}},
		Hex:    Coordinate{X: 2, Y: 2},
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
		ActivePlayer: playerId,
		PlayerCash:   map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState)
	err := performer.attemptTownPlacement(&TownPlacement{
		Tracks: []Direction{SOUTH_EAST},
		Hex:    Coordinate{X: 0, Y: 0},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_EAST, NORTH_WEST}},
		Hex:    Coordinate{X: 0, Y: 1},
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
		ActivePlayer: playerId,
		PlayerCash:   map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState)
	err := performer.attemptTownPlacement(&TownPlacement{
		Tracks: []Direction{SOUTH_EAST},
		Hex:    Coordinate{X: 0, Y: 0},
	})
	require.NoError(t, err)
	err = performer.attemptTrackPlacement(&TrackPlacement{
		Tracks: [][2]Direction{{NORTH_WEST, NORTH_EAST}},
		Hex:    Coordinate{X: 0, Y: 1},
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
		ActivePlayer: playerId,
		PlayerCash:   map[string]int{playerId: 10},
	}
	performer := newBuildActionPerformer(theMap, gameState)
	err := performer.attemptTownPlacement(&TownPlacement{
		Tracks: []Direction{SOUTH_EAST},
		Hex:    Coordinate{X: 0, Y: 0},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(gameState.Links))
	link := gameState.Links[0]
	assert.Equal(t, true, link.Complete)
	assert.Equal(t, playerId, link.Owner)
	assert.Equal(t, Coordinate{X: 0, Y: 0}, link.SourceHex)
	assert.Equal(t, []Direction{SOUTH_EAST}, link.Steps)
}
