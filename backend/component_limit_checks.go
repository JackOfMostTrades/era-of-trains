package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"net/http"
)

func checkTownMarkerLimit(mapState [][]*TileState) error {
	townMarkerCount := 0
	for y := 0; y < len(mapState); y++ {
		for x := 0; x < len(mapState[y]); x++ {
			ts := mapState[y][x]
			if ts.HasTown && !ts.IsCity && (len(ts.Routes) == 2 || len(ts.Routes) == 4) {
				townMarkerCount += 1
			}
		}
	}
	if townMarkerCount > 8 {
		return &HttpError{"Limit of town markers (8) is exceeded", http.StatusBadRequest}
	}
	return nil
}

type trackTile int

const (
	// Simple
	STRAIGHT_TRACK_TILE trackTile = iota + 1
	SHARP_CURVE_TRACK_TILE
	GENTLE_CURVE_TRACK_TILE

	// Complex crossing
	BOW_AND_ARROW_TRACK_TILE
	TWO_GENTLE_TRACK_TILE
	TWO_STRAIGHT_TRACK_TILE

	// Complex coexist
	BASEBALL_TRACK_TILE
	LEFT_GENTLE_AND_SHARP_TRACK_TILE
	RIGHT_GENTRLE_AND_SHARP_TRACK_TILE
	STRAIGHT_AND_SHARP_TRACK_TILE
)

var allTrackTiles []trackTile = []trackTile{
	STRAIGHT_TRACK_TILE, SHARP_CURVE_TRACK_TILE, GENTLE_CURVE_TRACK_TILE,
	BOW_AND_ARROW_TRACK_TILE, TWO_GENTLE_TRACK_TILE, TWO_STRAIGHT_TRACK_TILE,
	BASEBALL_TRACK_TILE, LEFT_GENTLE_AND_SHARP_TRACK_TILE, RIGHT_GENTRLE_AND_SHARP_TRACK_TILE, STRAIGHT_AND_SHARP_TRACK_TILE,
}

var trackTileRoutes = map[trackTile][][2]common.Direction{
	STRAIGHT_TRACK_TILE:     {{common.NORTH, common.SOUTH}},
	SHARP_CURVE_TRACK_TILE:  {{common.SOUTH_EAST, common.SOUTH}},
	GENTLE_CURVE_TRACK_TILE: {{common.NORTH_EAST, common.SOUTH}},

	BOW_AND_ARROW_TRACK_TILE: {{common.NORTH_EAST, common.SOUTH}, {common.SOUTH_EAST, common.NORTH_WEST}},
	TWO_GENTLE_TRACK_TILE:    {{common.NORTH, common.SOUTH_EAST}, {common.NORTH_EAST, common.SOUTH}},
	TWO_STRAIGHT_TRACK_TILE:  {{common.NORTH_EAST, common.SOUTH_WEST}, {common.SOUTH_EAST, common.NORTH_WEST}},

	BASEBALL_TRACK_TILE:                {{common.NORTH, common.SOUTH_WEST}, {common.NORTH_EAST, common.SOUTH}},
	LEFT_GENTLE_AND_SHARP_TRACK_TILE:   {{common.NORTH, common.SOUTH_EAST}, {common.SOUTH_WEST, common.NORTH_WEST}},
	RIGHT_GENTRLE_AND_SHARP_TRACK_TILE: {{common.NORTH, common.SOUTH_WEST}, {common.NORTH_EAST, common.SOUTH_EAST}},
	STRAIGHT_AND_SHARP_TRACK_TILE:      {{common.NORTH, common.SOUTH}, {common.SOUTH_WEST, common.NORTH_WEST}},
}

func routesEqual(a [][2]common.Direction, b []Route) bool {
	if len(a) != len(b) {
		return false
	}
	for _, trackA := range a {
		found := false
		for _, trackB := range b {
			if (trackA[0] == trackB.Left && trackA[1] == trackB.Right) ||
				(trackA[0] == trackB.Right && trackA[1] == trackB.Left) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func getTrackTileForRoutes(routes []Route) trackTile {
	for _, tile := range allTrackTiles {
		tileRoutes := trackTileRoutes[tile]
		for rotation := 0; rotation < 6; rotation++ {
			rotatedRoutes := make([][2]common.Direction, 0, len(tileRoutes))
			for _, route := range tileRoutes {
				rotatedRoutes = append(rotatedRoutes, [2]common.Direction{
					common.Direction((int(route[0]) + rotation) % 6), common.Direction((int(route[1]) + rotation) % 6),
				})
			}
			if routesEqual(rotatedRoutes, routes) {
				return tile
			}
		}
	}
	return 0
}

func checkTrackTileLimit(mapState [][]*TileState) error {
	componentCount := map[trackTile]int{
		STRAIGHT_TRACK_TILE:     48,
		SHARP_CURVE_TRACK_TILE:  7,
		GENTLE_CURVE_TRACK_TILE: 55,

		BOW_AND_ARROW_TRACK_TILE: 4,
		TWO_GENTLE_TRACK_TILE:    3,
		TWO_STRAIGHT_TRACK_TILE:  4,

		BASEBALL_TRACK_TILE:                1,
		LEFT_GENTLE_AND_SHARP_TRACK_TILE:   1,
		RIGHT_GENTRLE_AND_SHARP_TRACK_TILE: 1,
		STRAIGHT_AND_SHARP_TRACK_TILE:      1,
	}

	for y := 0; y < len(mapState); y++ {
		for x := 0; x < len(mapState[y]); x++ {
			ts := mapState[y][x]
			if ts.HasTown || ts.IsCity || len(ts.Routes) == 0 {
				continue
			}
			tile := getTrackTileForRoutes(ts.Routes)
			if tile == 0 {
				return fmt.Errorf("failed to identify track tile type for hex (%d,%d)", x, y)
			}
			if componentCount[tile] == 0 {
				return fmt.Errorf("ran out of track tiles for tile type %d", tile)
			}
			componentCount[tile] -= 1
		}
	}

	return nil
}
