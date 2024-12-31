package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
)

type DeliveryGraphLink struct {
	player      string
	destination common.Coordinate
}

type DeliveryGraph struct {
	hexToDirectionToLink map[common.Coordinate]map[common.Direction]DeliveryGraphLink
}

func applyDirection(coord common.Coordinate, direction common.Direction) common.Coordinate {
	switch direction {
	case common.NORTH:
		return common.Coordinate{X: coord.X, Y: coord.Y - 2}
	case common.NORTH_EAST:
		if (coord.Y % 2) == 0 {
			return common.Coordinate{X: coord.X, Y: coord.Y - 1}
		} else {
			return common.Coordinate{X: coord.X + 1, Y: coord.Y - 1}
		}
	case common.SOUTH_EAST:
		if (coord.Y % 2) == 0 {
			return common.Coordinate{X: coord.X, Y: coord.Y + 1}
		} else {
			return common.Coordinate{X: coord.X + 1, Y: coord.Y + 1}
		}
	case common.SOUTH:
		return common.Coordinate{X: coord.X, Y: coord.Y + 2}
	case common.SOUTH_WEST:
		if (coord.Y % 2) == 0 {
			return common.Coordinate{X: coord.X - 1, Y: coord.Y + 1}
		} else {
			return common.Coordinate{X: coord.X, Y: coord.Y + 1}
		}
	case common.NORTH_WEST:
		if (coord.Y % 2) == 0 {
			return common.Coordinate{X: coord.X - 1, Y: coord.Y - 1}
		} else {
			return common.Coordinate{X: coord.X, Y: coord.Y - 1}
		}
	}
	panic(fmt.Errorf("unhandled direction: %v", direction))
}

func computeDeliveryGraph(gameState *common.GameState, gameMap maps.GameMap) *DeliveryGraph {
	hexToDirectionToLink := make(map[common.Coordinate]map[common.Direction]DeliveryGraphLink)
	for _, link := range gameState.Links {
		if !link.Complete {
			continue
		}

		src := link.SourceHex
		dest := src
		for _, step := range link.Steps {
			if teleportDest, _ := gameMap.GetTeleportLink(dest, step); teleportDest != nil {
				dest = *teleportDest
			} else {
				dest = applyDirection(dest, step)
			}
		}
		if _, ok := hexToDirectionToLink[src]; !ok {
			hexToDirectionToLink[src] = make(map[common.Direction]DeliveryGraphLink)
		}
		if _, ok := hexToDirectionToLink[dest]; !ok {
			hexToDirectionToLink[dest] = make(map[common.Direction]DeliveryGraphLink)
		}

		hexToDirectionToLink[src][link.Steps[0]] = DeliveryGraphLink{
			player:      link.Owner,
			destination: dest,
		}
		hexToDirectionToLink[dest][link.Steps[len(link.Steps)-1].Opposite()] = DeliveryGraphLink{
			player:      link.Owner,
			destination: src,
		}
	}
	return &DeliveryGraph{hexToDirectionToLink: hexToDirectionToLink}
}
