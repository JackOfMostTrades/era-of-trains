package main

import (
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"slices"
)

func (handler *confirmMoveHandler) checkRouteConnections() error {
	// This runs a DFS style algorithm to mark all links (and recursively everything they connect to) as valid, starting from cities
	// Afterward, we check all the player's links to make sure they have been marked as valid or else return an error

	validLinks := make(map[*common.Link]bool)
	var exploreQueue []common.Coordinate
	var seenHexes []common.Coordinate

	// Enumerate all city hexes to be explored
	for y := 0; y < handler.gameMap.GetHeight(); y++ {
		for x := 0; x < handler.gameMap.GetWidth(); x++ {
			hex := common.Coordinate{X: x, Y: y}
			if handler.gameMap.GetHexType(hex) == maps.CITY_HEX_TYPE {
				exploreQueue = append(exploreQueue, hex)
			}
		}
	}
	for _, urb := range handler.gameState.Urbanizations {
		exploreQueue = append(exploreQueue, urb.Hex)
	}

	// Explore from every hex (not yet explored) until the queue is empty.
	for len(exploreQueue) > 0 {
		hex := exploreQueue[len(exploreQueue)-1]
		exploreQueue = exploreQueue[0 : len(exploreQueue)-1]
		if slices.Index(seenHexes, hex) != -1 {
			continue
		}
		seenHexes = append(seenHexes, hex)

		for _, link := range handler.gameState.Links {
			if link.Owner != handler.activePlayer {
				continue
			}

			startHex := link.SourceHex
			endHex := startHex
			for _, step := range link.Steps {
				endHex = applyMapDirection(handler.gameMap, handler.gameState, endHex, step)
			}

			if startHex == hex {
				validLinks[link] = true
				if link.Complete {
					exploreQueue = append(exploreQueue, endHex)
				}
			}
			if link.Complete && endHex == hex {
				validLinks[link] = true
				exploreQueue = append(exploreQueue, startHex)
			}
		}
	}

	// Check if any of the player's links has not been marked as valid
	for _, link := range handler.gameState.Links {
		if link.Owner != handler.activePlayer {
			continue
		}
		if isValid, ok := validLinks[link]; !ok || !isValid {
			return invalidMoveErr("all of a player's links must trace back over a player's track to a city")
		}
	}

	return nil
}

func (handler *confirmMoveHandler) checkLoopingConnections() error {
	// This checks every completed link to verify it does not start and end at the same hex, since links are not
	// allowed to connect directly back to the same city/town as where they started

	for _, link := range handler.gameState.Links {
		if !link.Complete {
			continue
		}

		endHex := link.SourceHex
		for _, step := range link.Steps {
			if teleportDest, _ := handler.gameMap.GetTeleportLink(handler.gameState, endHex, step); teleportDest != nil {
				endHex = *teleportDest
			} else {
				endHex = applyDirection(endHex, step)
			}
		}

		if link.SourceHex.Equals(endHex) {
			return invalidMoveErr("individual links are not allowed to start and end at the same town/city")
		}
	}

	return nil
}
