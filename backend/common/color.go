package common

import "strconv"

type Color int

const (
	NONE_COLOR Color = iota
	BLACK
	RED
	YELLOW
	BLUE
	PURPLE
	WHITE
)

func (c Color) String() string {
	switch c {
	case NONE_COLOR:
		return "NONE"
	case BLACK:
		return "BLACK"
	case RED:
		return "RED"
	case YELLOW:
		return "YELLOW"
	case BLUE:
		return "BLUE"
	case PURPLE:
		return "PURPLE"
	case WHITE:
		return "WHITE"
	}
	return strconv.Itoa(int(c))
}
