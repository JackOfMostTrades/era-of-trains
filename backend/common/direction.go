package common

import "fmt"

type Direction int

const (
	NORTH Direction = iota
	NORTH_EAST
	SOUTH_EAST
	SOUTH
	SOUTH_WEST
	NORTH_WEST
)

var ALL_DIRECTIONS = []Direction{NORTH, NORTH_EAST, SOUTH_EAST, SOUTH, SOUTH_WEST, NORTH_WEST}

func (d Direction) Opposite() Direction {
	switch d {
	case NORTH:
		return SOUTH
	case NORTH_EAST:
		return SOUTH_WEST
	case SOUTH_EAST:
		return NORTH_WEST
	case SOUTH:
		return NORTH
	case SOUTH_WEST:
		return NORTH_EAST
	case NORTH_WEST:
		return SOUTH_EAST
	}
	panic(fmt.Errorf("unhandeled direction: %v", d))
}
