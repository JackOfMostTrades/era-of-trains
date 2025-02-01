package main

import (
	"iter"

	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
)

type Dangler struct {
	from common.Direction
	link *common.Link
}

type Route struct {
	// Default value and meaningless if there is a town. Otherwise, one edge of the route
	Left common.Direction
	// Other edge of the route
	Right common.Direction
	// What link is this a part of?
	Link *common.Link
}

type TileState struct {
	routes   []*Route
	danglers []*Dangler
	isTown   bool
	isCity   bool
}

type MapState struct {
	tileState [][]*TileState
}

func newMapState(gameMap maps.GameMap, gameState *common.GameState) *MapState {
	ts := make([][]*TileState, gameMap.GetHeight())
	for y := 0; y < gameMap.GetHeight(); y++ {
		ts[y] = make([]*TileState, gameMap.GetWidth())
		for x := 0; x < gameMap.GetWidth(); x++ {
			hexType := gameMap.GetHexType(common.Coordinate{X: x, Y: y})
			ts[y][x] = &TileState{
				routes:   nil,
				danglers: nil,
				isTown:   hexType == maps.TOWN_HEX_TYPE,
				isCity:   hexType == maps.CITY_HEX_TYPE,
			}
		}
	}

	if gameState != nil {
		for _, urb := range gameState.Urbanizations {
			ts[urb.Hex.Y][urb.Hex.X].isCity = true
		}
		for _, link := range gameState.Links {
			hex := link.SourceHex
			ts[hex.Y][hex.X].routes = append(ts[hex.Y][hex.X].routes, &Route{
				Left:  link.Steps[0],
				Right: link.Steps[0],
				Link:  link,
			})

			for i := 1; i < len(link.Steps); i++ {
				hex = applyDirection(hex, link.Steps[i-1])
				ts[hex.Y][hex.X].routes = append(ts[hex.Y][hex.X].routes, &Route{
					Left:  link.Steps[i-1].Opposite(),
					Right: link.Steps[i],
					Link:  link,
				})
			}

			if !link.Complete && len(link.Steps) > 1 {
				ts[hex.Y][hex.X].danglers = append(ts[hex.Y][hex.X].danglers, &Dangler{
					from: link.Steps[len(link.Steps)-2].Opposite(),
					link: link,
				})
			}

			hex = applyDirection(hex, link.Steps[len(link.Steps)-1])
			if ts[hex.Y][hex.X].isTown && link.Complete {
				dir := link.Steps[len(link.Steps)-1].Opposite()
				ts[hex.Y][hex.X].routes = append(ts[hex.Y][hex.X].routes, &Route{
					Left:  dir,
					Right: dir,
					Link:  link,
				})
			}
		}
	}

	return &MapState{
		tileState: ts,
	}
}

func (ms *MapState) GetTileState(hex common.Coordinate) *TileState {
	if hex.Y < 0 || hex.Y >= len(ms.tileState) || hex.X < 0 || hex.X >= len(ms.tileState[hex.Y]) {
		return nil
	}
	return ms.tileState[hex.Y][hex.X]
}

func (ms *MapState) GetAllTileState() iter.Seq2[common.Coordinate, *TileState] {
	return func(yield func(common.Coordinate, *TileState) bool) {
		for y := 0; y < len(ms.tileState); y++ {
			for x := 0; x < len(ms.tileState[y]); x++ {
				if !yield(common.Coordinate{X: x, Y: y}, ms.tileState[y][x]) {
					return
				}
			}
		}
	}
}
