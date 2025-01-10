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

func applyMapDirection(gameMap maps.GameMap, gameState *common.GameState, hex common.Coordinate, direction common.Direction) common.Coordinate {
	if teleportDest, _ := gameMap.GetTeleportLink(gameState, hex, direction); teleportDest != nil {
		return *teleportDest
	} else {
		return applyDirection(hex, direction)
	}
}

func computeDeliveryGraph(gameState *common.GameState, gameMap maps.GameMap) *DeliveryGraph {
	hexToDirectionToLink := make(map[common.Coordinate]map[common.Direction]DeliveryGraphLink)
	for _, link := range gameState.Links {
		if !link.Complete {
			continue
		}

		src := link.SourceHex
		dest := src
		var lastReverseDirection common.Direction
		for _, step := range link.Steps {
			if teleportDest, teleportDirection := gameMap.GetTeleportLink(gameState, dest, step); teleportDest != nil {
				dest = *teleportDest
				lastReverseDirection = teleportDirection
			} else {
				dest = applyDirection(dest, step)
				lastReverseDirection = step.Opposite()
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
		hexToDirectionToLink[dest][lastReverseDirection] = DeliveryGraphLink{
			player:      link.Owner,
			destination: src,
		}
	}
	return &DeliveryGraph{hexToDirectionToLink: hexToDirectionToLink}
}
