package main

import (
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/JackOfMostTrades/eot/backend/tiles"
	"slices"
)

type tilePlacer struct {
	performer *buildActionPerformer
	mapState  *MapState
}

func newTilePlacer(performer *buildActionPerformer) *tilePlacer {
	return &tilePlacer{
		performer: performer,
		mapState:  newMapState(performer.gameMap, performer.gameState),
	}
}

func (tp *tilePlacer) applyTrackTilePlacement(coordinate common.Coordinate,
	tile tiles.TrackTile,
	rotation int) error {

	activePlayer := tp.performer.activePlayer
	gameState := tp.performer.gameState

	ts := tp.mapState.GetTileState(coordinate)
	if ts.isTown || ts.isCity {
		return invalidMoveErr("cannot place track tile on city or town hex")
	}

	var newRoutes [][2]common.Direction
	for _, route := range tiles.GetRoutesForTile(tile) {
		rotatedRoute := [2]common.Direction{
			common.Direction((int(route[0]) + rotation) % 6),
			common.Direction((int(route[1]) + rotation) % 6),
		}
		newRoutes = append(newRoutes, rotatedRoute)
	}

	// Check that any routes that must be preserved are in fact preserved
	for _, oldRoute := range ts.routes {
		mustBePreserved := false
		// Any routes owned by another player must be preserved
		if oldRoute.Link.Owner != "" && oldRoute.Link.Owner != activePlayer {
			mustBePreserved = true
		}
		// Any routes that are not a dangler must be preserved
		isDangler := false
		for _, dangler := range ts.danglers {
			if dangler.from == oldRoute.Left || dangler.from == oldRoute.Right {
				isDangler = true
				break
			}
		}
		if !isDangler {
			mustBePreserved = true
		}

		// Check that it is preserved if it must be preserved.
		if mustBePreserved {
			isPreserved := false
			for _, newRoute := range newRoutes {
				if (newRoute[0] == oldRoute.Left && newRoute[1] == oldRoute.Right) ||
					(newRoute[0] == oldRoute.Right && newRoute[1] == oldRoute.Left) {
					isPreserved = true
					break
				}
			}
			if !isPreserved {
				return invalidMoveErr("routes owned by other players must be preserved")
			}
		}
	}

	// Check that there is a track from each dangler direction
	for _, dangler := range ts.danglers {
		hasExtension := false
		for _, newRoute := range newRoutes {
			if newRoute[0] == dangler.from || newRoute[1] == dangler.from {
				hasExtension = true
			}
		}
		if !hasExtension {
			return invalidMoveErr("redirected dangling track must preserve origin direction")
		}
	}

	// Every exit must lead to a passable terrain
	for _, newRoute := range newRoutes {
		for _, dir := range newRoute {
			newHex := applyDirection(coordinate, dir)
			if newHex.X < 0 || newHex.Y < 0 ||
				newHex.X >= tp.performer.gameMap.GetWidth() || newHex.Y >= tp.performer.gameMap.GetHeight() {
				return invalidMoveErr("track cannot run off the edge of the board")
			}

			newHexType := tp.performer.gameMap.GetHexType(newHex)
			if newHexType == maps.WATER_HEX_TYPE || newHexType == maps.OFFBOARD_HEX_TYPE {
				return invalidMoveErr("track cannot lead to water or offboard spaces")
			}
		}
	}

	for _, newRoute := range newRoutes {
		isOldRoute := false
		for _, oldRoute := range ts.routes {
			if (newRoute[0] == oldRoute.Left && newRoute[1] == oldRoute.Right) ||
				(newRoute[0] == oldRoute.Right && newRoute[1] == oldRoute.Left) {
				isOldRoute = true
				break
			}
		}
		if !isOldRoute {
			var leftDangler *Dangler
			var rightDangler *Dangler
			for _, dangler := range ts.danglers {
				if dangler.from == newRoute[0] {
					leftDangler = dangler
				}
				if dangler.from == newRoute[1] {
					rightDangler = dangler
				}
			}

			// Drop the last step of the danglers since they are redirected by this track
			if leftDangler != nil {
				leftDangler.link.Steps = leftDangler.link.Steps[:len(leftDangler.link.Steps)-1]
			}
			if rightDangler != nil {
				rightDangler.link.Steps = rightDangler.link.Steps[:len(rightDangler.link.Steps)-1]
			}

			// Add this route. See if this attaches to any existing link
			var attachingLink *common.Link
			var remainingDirection common.Direction

			if link := tp.getAdjoiningLink(coordinate, newRoute[0]); link != nil {
				attachingLink = link
				remainingDirection = newRoute[1]
			} else if link := tp.getAdjoiningLink(coordinate, newRoute[1]); link != nil {
				attachingLink = link
				remainingDirection = newRoute[0]
			}

			if attachingLink == nil {
				// Create a new route. The left or right side must be a city since it is not connecting to track
				leftHex := applyDirection(coordinate, newRoute[0])
				rightHex := applyDirection(coordinate, newRoute[1])
				if tp.mapState.GetTileState(leftHex).isCity {
					newLink := &common.Link{
						SourceHex: leftHex,
						Steps:     []common.Direction{newRoute[0].Opposite(), newRoute[1]},
						Owner:     activePlayer,
						Complete:  tp.mapState.GetTileState(rightHex).isCity,
					}
					gameState.Links = append(gameState.Links, newLink)
					tp.performer.extendedLinks[newLink] = true
				} else if tp.mapState.GetTileState(rightHex).isCity {
					newLink := &common.Link{
						SourceHex: rightHex,
						Steps:     []common.Direction{newRoute[1].Opposite(), newRoute[0]},
						Owner:     activePlayer,
						Complete:  false,
					}
					gameState.Links = append(gameState.Links, newLink)
					tp.performer.extendedLinks[newLink] = true
				} else {
					return invalidMoveErr("track must connect to other track or a city")
				}
			} else {
				if attachingLink.Owner != "" && attachingLink.Owner != activePlayer {
					return invalidMoveErr("cannot join to another player's links")
				}
				attachingLink.Owner = activePlayer
				attachingLink.Steps = append(attachingLink.Steps, remainingDirection)
				tp.performer.extendedLinks[attachingLink] = true

				// If the other side is also a link, we need to combine the two links
				if rightLink := tp.getAdjoiningLink(coordinate, remainingDirection); rightLink != nil {
					if rightLink.Owner != "" && rightLink.Owner != activePlayer {
						return invalidMoveErr("cannot join to another player's links")
					}
					for i := len(rightLink.Steps) - 2; i >= 0; i-- {
						attachingLink.Steps = append(attachingLink.Steps, rightLink.Steps[i].Opposite())
					}
					// Connecting two sides together, so the link is now complete
					attachingLink.Complete = true
					gameState.Links = DeleteFromSliceUnordered(slices.Index(gameState.Links, rightLink), gameState.Links)
				} else {
					// If the other side is a city, we need to mark the link as complete
					rightHex := applyDirection(coordinate, remainingDirection)
					if tp.mapState.GetTileState(rightHex).isCity {
						attachingLink.Complete = true
					}
				}
			}
		}
	}

	return nil
}

func (tp *tilePlacer) getAdjoiningLink(fromHex common.Coordinate, direction common.Direction) *common.Link {
	hex := applyDirection(fromHex, direction)
	ts := tp.mapState.GetTileState(hex)
	opp := direction.Opposite()
	for _, route := range ts.routes {
		if route.Left == opp || route.Right == opp {
			return route.Link
		}
	}
	return nil
}
