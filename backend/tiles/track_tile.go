package tiles

import "github.com/JackOfMostTrades/eot/backend/common"

type TrackTile int

const (
	// Simple
	STRAIGHT_TRACK_TILE TrackTile = iota + 1
	SHARP_CURVE_TRACK_TILE
	GENTLE_CURVE_TRACK_TILE

	// Complex crossing
	BOW_AND_ARROW_TRACK_TILE
	TWO_GENTLE_TRACK_TILE
	TWO_STRAIGHT_TRACK_TILE

	// Complex coexist
	BASEBALL_TRACK_TILE
	LEFT_GENTLE_AND_SHARP_TRACK_TILE
	RIGHT_GENTLE_AND_SHARP_TRACK_TILE
	STRAIGHT_AND_SHARP_TRACK_TILE
)

var AllTrackTiles []TrackTile = []TrackTile{
	STRAIGHT_TRACK_TILE, SHARP_CURVE_TRACK_TILE, GENTLE_CURVE_TRACK_TILE,
	BOW_AND_ARROW_TRACK_TILE, TWO_GENTLE_TRACK_TILE, TWO_STRAIGHT_TRACK_TILE,
	BASEBALL_TRACK_TILE, LEFT_GENTLE_AND_SHARP_TRACK_TILE, RIGHT_GENTLE_AND_SHARP_TRACK_TILE, STRAIGHT_AND_SHARP_TRACK_TILE,
}

var trackTileRoutes = map[TrackTile][][2]common.Direction{
	STRAIGHT_TRACK_TILE:     {{common.NORTH, common.SOUTH}},
	SHARP_CURVE_TRACK_TILE:  {{common.SOUTH_EAST, common.SOUTH}},
	GENTLE_CURVE_TRACK_TILE: {{common.NORTH_EAST, common.SOUTH}},

	BOW_AND_ARROW_TRACK_TILE: {{common.NORTH_EAST, common.SOUTH}, {common.SOUTH_EAST, common.NORTH_WEST}},
	TWO_GENTLE_TRACK_TILE:    {{common.NORTH, common.SOUTH_EAST}, {common.NORTH_EAST, common.SOUTH}},
	TWO_STRAIGHT_TRACK_TILE:  {{common.NORTH_EAST, common.SOUTH_WEST}, {common.SOUTH_EAST, common.NORTH_WEST}},

	BASEBALL_TRACK_TILE:               {{common.NORTH, common.SOUTH_WEST}, {common.NORTH_EAST, common.SOUTH}},
	LEFT_GENTLE_AND_SHARP_TRACK_TILE:  {{common.NORTH, common.SOUTH_EAST}, {common.SOUTH_WEST, common.NORTH_WEST}},
	RIGHT_GENTLE_AND_SHARP_TRACK_TILE: {{common.NORTH, common.SOUTH_WEST}, {common.NORTH_EAST, common.SOUTH_EAST}},
	STRAIGHT_AND_SHARP_TRACK_TILE:     {{common.NORTH, common.SOUTH}, {common.SOUTH_WEST, common.NORTH_WEST}},
}

func GetRoutesForTile(tile TrackTile) [][2]common.Direction {
	return trackTileRoutes[tile]
}
