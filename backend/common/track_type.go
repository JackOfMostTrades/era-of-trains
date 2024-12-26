package common

type TrackType int

const (
	SIMPLE_TRACK_TYPE TrackType = iota
	COMPLEX_CROSSING_TRACK_TYPE
	COMPLEX_COEXISTING_TRACK_TYPE
)

func routesEqual(a [][2]Direction, b [][2]Direction) bool {
	if len(a) != len(b) {
		return false
	}
	for _, trackA := range a {
		found := false
		for _, trackB := range b {
			if (trackA[0] == trackB[0] && trackA[1] == trackB[1]) ||
				(trackA[0] == trackB[1] && trackA[1] == trackB[0]) {
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

func GetTrackType(routes [][2]Direction) TrackType {
	if len(routes) < 2 {
		return SIMPLE_TRACK_TYPE
	}

	complexCrossingTiles := [][][2]Direction{
		// X
		{{NORTH, SOUTH}, {SOUTH_WEST, NORTH_EAST}},
		// Gentle X
		{{NORTH, SOUTH_EAST}, {NORTH_EAST, SOUTH}},
		// Bow and arrow
		{{NORTH, SOUTH}, {SOUTH_WEST, SOUTH_EAST}},
	}
	for _, tile := range complexCrossingTiles {
		for rotation := 0; rotation < 6; rotation++ {
			rotatedRoutes := make([][2]Direction, 0, len(tile))
			for _, route := range tile {
				rotatedRoutes = append(rotatedRoutes, [2]Direction{
					Direction((int(route[0]) + rotation) % 6), Direction((int(route[1]) + rotation) % 6),
				})
			}
			if routesEqual(rotatedRoutes, routes) {
				return COMPLEX_CROSSING_TRACK_TYPE
			}
		}
	}

	return COMPLEX_COEXISTING_TRACK_TYPE
}
