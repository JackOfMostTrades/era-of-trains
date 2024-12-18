package main

import "fmt"

type DeliveryGraphLink struct {
	player      string
	destination Coordinate
}

type DeliveryGraph struct {
	hexToDirectionToLink map[Coordinate]map[Direction]DeliveryGraphLink
}

func applyDirection(coord Coordinate, direction Direction) Coordinate {
	switch direction {
	case NORTH:
		return Coordinate{X: coord.X, Y: coord.Y - 2}
	case NORTH_EAST:
		if (coord.Y % 2) == 0 {
			return Coordinate{X: coord.X, Y: coord.Y - 1}
		} else {
			return Coordinate{X: coord.X + 1, Y: coord.Y - 1}
		}
	case SOUTH_EAST:
		if (coord.Y % 2) == 0 {
			return Coordinate{X: coord.X, Y: coord.Y + 1}
		} else {
			return Coordinate{X: coord.X + 1, Y: coord.Y + 1}
		}
	case SOUTH:
		return Coordinate{X: coord.X, Y: coord.Y + 2}
	case SOUTH_WEST:
		if (coord.Y % 2) == 0 {
			return Coordinate{X: coord.X - 1, Y: coord.Y + 1}
		} else {
			return Coordinate{X: coord.X, Y: coord.Y + 1}
		}
	case NORTH_WEST:
		if (coord.Y % 2) == 0 {
			return Coordinate{X: coord.X - 1, Y: coord.Y - 1}
		} else {
			return Coordinate{X: coord.X, Y: coord.Y - 1}
		}
	}
	panic(fmt.Errorf("unhandled direction: %v", direction))
}

func (gameState *GameState) computeDeliveryGraph() *DeliveryGraph {
	hexToDirectionToLink := make(map[Coordinate]map[Direction]DeliveryGraphLink)
	for _, link := range gameState.Links {
		if !link.Complete {
			continue
		}

		src := link.SourceHex
		dest := src
		for _, step := range link.Steps {
			dest = applyDirection(dest, step)
		}
		if _, ok := hexToDirectionToLink[src]; !ok {
			hexToDirectionToLink[src] = make(map[Direction]DeliveryGraphLink)
		}
		if _, ok := hexToDirectionToLink[dest]; !ok {
			hexToDirectionToLink[dest] = make(map[Direction]DeliveryGraphLink)
		}

		hexToDirectionToLink[src][link.Steps[0]] = DeliveryGraphLink{
			player:      link.Owner,
			destination: dest,
		}
		hexToDirectionToLink[dest][link.Steps[len(link.Steps)-1].opposite()] = DeliveryGraphLink{
			player:      link.Owner,
			destination: src,
		}
	}
	return &DeliveryGraph{hexToDirectionToLink: hexToDirectionToLink}
}
