package main

import "net/http"

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

func checkTrackTileLimit(mapState [][]*TileState) error {
	return nil
}
