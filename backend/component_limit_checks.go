package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/tiles"
	"net/http"
)

func checkTownMarkerLimit(mapState *MapState) error {
	townMarkerCount := 0
	for _, ts := range mapState.GetAllTileState() {
		if ts.isTown && !ts.isCity && (len(ts.routes) == 2 || len(ts.routes) == 4) {
			townMarkerCount += 1
		}
	}
	if townMarkerCount > 8 {
		return &HttpError{"Limit of town markers (8) is exceeded", http.StatusBadRequest}
	}
	return nil
}

func routesEqual(a [][2]common.Direction, b []*Route) bool {
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

func getTrackTileForRoutes(routes []*Route) tiles.TrackTile {
	for _, tile := range tiles.AllTrackTiles {
		tileRoutes := tiles.GetRoutesForTile(tile)
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

func checkTrackTileLimit(mapState *MapState) error {
	componentCount := map[tiles.TrackTile]int{
		tiles.STRAIGHT_TRACK_TILE:     48,
		tiles.SHARP_CURVE_TRACK_TILE:  7,
		tiles.GENTLE_CURVE_TRACK_TILE: 55,

		tiles.BOW_AND_ARROW_TRACK_TILE: 4,
		tiles.TWO_GENTLE_TRACK_TILE:    3,
		tiles.TWO_STRAIGHT_TRACK_TILE:  4,

		tiles.BASEBALL_TRACK_TILE:               1,
		tiles.LEFT_GENTLE_AND_SHARP_TRACK_TILE:  1,
		tiles.RIGHT_GENTLE_AND_SHARP_TRACK_TILE: 1,
		tiles.STRAIGHT_AND_SHARP_TRACK_TILE:     1,
	}

	for _, ts := range mapState.GetAllTileState() {
		if ts.isTown || ts.isCity || len(ts.routes) == 0 {
			continue
		}
		tile := getTrackTileForRoutes(ts.routes)
		if tile == 0 {
			return fmt.Errorf("failed to identify track tile type for hex")
		}
		if componentCount[tile] == 0 {
			return invalidMoveErr("ran out of track tiles for tile type %d", tile)
		}
		componentCount[tile] -= 1
	}

	return nil
}
