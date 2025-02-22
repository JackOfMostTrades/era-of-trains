package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/api"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/JackOfMostTrades/eot/backend/tiles"
	"slices"
)

func (performer *buildActionPerformer) attemptTownPlacement(hex common.Coordinate, townPlacement *api.TownPlacement) error {
	mapState := newMapState(performer.gameMap, performer.gameState)
	ts := mapState.GetTileState(hex)

	if !ts.isTown {
		return invalidMoveErr("cannot build town track on a non-town hex")
	}
	if len(townPlacement.Track) > 4 {
		return invalidMoveErr("cannot build more than four tracks on a town hex")
	}
	// Verify that all existing routes are preserved
	for _, existingRoute := range ts.routes {
		if slices.Index(townPlacement.Track, existingRoute.Right) == -1 {
			return invalidMoveErr("placement of town tile must preserve all existing track")
		}
	}

	for _, direction := range townPlacement.Track {
		// If this is existing track, do nothing
		isExisting := false
		for _, existingRoute := range ts.routes {
			if existingRoute.Right == direction {
				isExisting = true
				break
			}
		}
		if isExisting {
			continue
		}

		// If it hits a stop add a new link for the player
		// If it hits player existing track, then mark it as complete
		// If it hits nothing, add an incomplete new link for the player

		nextHex := applyDirection(hex, direction)
		next := mapState.GetTileState(nextHex)
		var link *common.Link
		if next.isCity {
			link = &common.Link{
				SourceHex: hex,
				Owner:     performer.activePlayer,
				Steps:     []common.Direction{direction},
				Complete:  true,
			}
			performer.gameState.Links = append(performer.gameState.Links, link)
		} else {
			isJoiningRoute := false
			for _, route := range next.routes {
				if route.Left == direction.Opposite() || route.Right == direction.Opposite() {
					// Check that we are not joining into a different player's track
					if route.Link.Owner != "" && route.Link.Owner != performer.activePlayer {
						return invalidMoveErr("cannot build track that connects to another player's track")
					}

					link = route.Link
					route.Link.Complete = true
					route.Link.Owner = performer.activePlayer
					isJoiningRoute = true
					break
				}
			}
			if !isJoiningRoute {
				link = &common.Link{
					SourceHex: hex,
					Owner:     performer.activePlayer,
					Steps:     []common.Direction{direction},
					Complete:  false,
				}
				performer.gameState.Links = append(performer.gameState.Links, link)
				performer.extendedLinks[link] = true
			}
		}
	}

	return nil
}

func (performer *buildActionPerformer) attemptTeleportLinkPlacement(hex common.Coordinate, teleportLinkPlacement *api.TeleportLinkPlacement) error {
	direction := teleportLinkPlacement.Track

	otherHex, otherDirection := performer.gameMap.GetTeleportLink(performer.gameState, hex, direction)
	if otherHex == nil {
		return invalidMoveErr("cannot place teleport link at %s in direction %d",
			renderHexCoordinate(hex), direction)
	}

	// Check all steps of all links and validate both sides of the teleport
	for _, playerLink := range performer.gameState.Links {
		linkHex := playerLink.SourceHex
		for _, step := range playerLink.Steps {
			if linkHex == hex && step == direction {
				return invalidMoveErr("another player has already built on this link")
			}
			if linkHex == *otherHex && step == otherDirection {
				return invalidMoveErr("another player has already built on this link")
			}
			linkHex = applyDirection(linkHex, step)
		}
	}

	performer.gameState.Links = append(performer.gameState.Links, &common.Link{
		SourceHex: hex,
		Steps:     []common.Direction{direction},
		Owner:     performer.activePlayer,
		Complete:  true,
	})

	return nil
}

func (performer *buildActionPerformer) determineTownBuildCost(hex common.Coordinate, townPlacement *api.TownPlacement) (int, error) {
	ts := newMapState(performer.gameMap, performer.gameState).GetTileState(hex)

	var cost int
	cost = performer.gameMap.GetTownBuildCost(performer.gameState, performer.activePlayer, hex, len(townPlacement.Track), len(ts.routes) != 0)
	return cost, nil
}

func (performer *buildActionPerformer) determineTrackBuildCost(hex common.Coordinate, trackPlacement *api.TrackPlacement) (int, error) {
	ts := newMapState(performer.gameMap, performer.gameState).GetTileState(hex)

	hexType := performer.gameMap.GetHexType(hex)
	cost, err := performer.gameMap.GetTrackBuildCost(performer.gameState, performer.activePlayer,
		hexType, hex, tiles.GetTrackType(trackPlacement.Tile), len(ts.routes) != 0)
	if err != nil {
		return 0, fmt.Errorf("failed to determine cost for placing track tile: %v", err)
	}

	return cost, nil
}

func (performer *buildActionPerformer) attemptTrackPlacement(hex common.Coordinate, trackPlacement *api.TrackPlacement) error {
	tilePlacer := newTilePlacer(performer)
	err := tilePlacer.applyTrackTilePlacement(hex,
		trackPlacement.Tile,
		trackPlacement.Rotation)
	if err != nil {
		return err
	}

	return nil
}

type buildActionPerformer struct {
	extendedLinks map[*common.Link]bool
	gameState     *common.GameState
	activePlayer  string
	gameMap       maps.GameMap
}

func newBuildActionPerformer(gameMap maps.GameMap, gameState *common.GameState, activePlayer string) *buildActionPerformer {

	performer := &buildActionPerformer{
		extendedLinks: make(map[*common.Link]bool),
		gameState:     gameState,
		activePlayer:  activePlayer,
		gameMap:       gameMap,
	}

	return performer
}

func (performer *buildActionPerformer) handleUrbanization(hex common.Coordinate, city int) error {
	gameState := performer.gameState
	if gameState.PlayerActions[performer.activePlayer] != common.URBANIZATION_SPECIAL_ACTION {
		return invalidMoveErr("cannot urbanize without special action")
	}
	if city < 0 || city >= 8 {
		return invalidMoveErr("invalid city: %d", city)
	}

	for _, existingUrb := range gameState.Urbanizations {
		if existingUrb.Hex == hex {
			return invalidMoveErr("cannot urbanize on top of existing urbanization")
		}
		if existingUrb.City == city {
			return invalidMoveErr("requested city has already been urbanized")
		}
	}
	if performer.gameMap.GetHexType(hex) != maps.TOWN_HEX_TYPE {
		return invalidMoveErr("must urbanize on town hex")
	}

	gameState.Urbanizations = append(gameState.Urbanizations, &common.Urbanization{
		Hex:  hex,
		City: city,
	})

	// Check if there is adjacent incomplete link that becomes completed by this build
	mapState := newMapState(performer.gameMap, performer.gameState)
	for _, direction := range common.ALL_DIRECTIONS {
		adjacentHex := applyDirection(hex, direction)
		ts := mapState.GetTileState(adjacentHex)
		if ts != nil {
			for _, route := range ts.routes {
				if route.Left == direction.Opposite() || route.Right == direction.Opposite() {
					route.Link.Complete = true
				}
			}
		}
	}

	// Any single-step incomplete links coming out of the town (stubs) get destroyed by urbanization
	for i := 0; i < len(gameState.Links); i++ {
		link := gameState.Links[i]
		if link.SourceHex.X == hex.X && link.SourceHex.Y == hex.Y &&
			!link.Complete && len(link.Steps) == 1 {
			gameState.Links = DeleteFromSliceUnordered(i, gameState.Links)
			i -= 1
		}
	}

	return nil
}

func countsForHexPlacement(step *api.BuildStep) bool {
	if step.Urbanization != nil || step.TownPlacement != nil || step.TrackPlacement != nil {
		return true
	}
	return false
}

func (handler *confirmMoveHandler) performBuildAction(buildAction *api.BuildAction) error {

	gameState := handler.gameState
	performer := newBuildActionPerformer(handler.gameMap, handler.gameState, handler.activePlayer)

	// Validate no repeated hexes in the steps
	for i := 0; i < len(buildAction.Steps); i++ {
		if !countsForHexPlacement(buildAction.Steps[i]) {
			continue
		}
		for j := i + 1; j < len(buildAction.Steps); j++ {
			if !countsForHexPlacement(buildAction.Steps[j]) {
				continue
			}

			if buildAction.Steps[i].Hex.Equals(buildAction.Steps[j].Hex) {
				return invalidMoveErr("cannot perform multiple builds on the same hex on the same turn")
			}
		}
	}

	nonUrbStepCount := 0
	for _, step := range buildAction.Steps {
		if step.Urbanization == nil {
			nonUrbStepCount += 1
		}
	}

	// Check the number of placements is valid
	placementLimit, err := handler.gameMap.GetBuildLimit(gameState, handler.activePlayer)
	if err != nil {
		return err
	}
	if nonUrbStepCount > placementLimit {
		return invalidMoveErr("cannot exceed track placement limit (%d)", placementLimit)
	}

	// Apply builds and count up the costs
	var costs []int
	for _, step := range buildAction.Steps {
		if step.Urbanization != nil {
			err := performer.handleUrbanization(step.Hex, *step.Urbanization)
			if err != nil {
				return err
			}
			handler.Log("%s urbanizes new city %c at %s",
				handler.ActivePlayerNick(), 'A'+*step.Urbanization, renderHexCoordinate(step.Hex))
		}
		if step.TownPlacement != nil {
			cost, err := performer.determineTownBuildCost(step.Hex, step.TownPlacement)
			if err != nil {
				return err
			}
			costs = append(costs, cost)
			err = performer.attemptTownPlacement(step.Hex, step.TownPlacement)
			if err != nil {
				return err
			}
			handler.Log("%s added track on town hex %s",
				handler.ActivePlayerNick(), renderHexCoordinate(step.Hex))
		}
		if step.TrackPlacement != nil {
			cost, err := performer.determineTrackBuildCost(step.Hex, step.TrackPlacement)
			if err != nil {
				return err
			}
			costs = append(costs, cost)
			err = performer.attemptTrackPlacement(step.Hex, step.TrackPlacement)
			if err != nil {
				return err
			}
			handler.Log("%s added track on hex %s",
				handler.ActivePlayerNick(), renderHexCoordinate(step.Hex))
		}
		if step.TeleportLinkPlacement != nil {
			cost := handler.gameMap.GetTeleportLinkBuildCost(gameState, handler.activePlayer,
				step.Hex, step.TeleportLinkPlacement.Track)
			if cost == 0 {
				return invalidMoveErr("invalid teleport link placement (no teleport link exists in the target hex/direction)")
			}
			costs = append(costs, cost)
			err := performer.attemptTeleportLinkPlacement(step.Hex, step.TeleportLinkPlacement)
			if err != nil {
				return err
			}
			handler.Log("%s added teleport link on hex %s",
				handler.ActivePlayerNick(), renderHexCoordinate(step.Hex))
		}
	}

	totalCost := handler.gameMap.GetTotalBuildCost(gameState, handler.activePlayer, costs)
	if totalCost > gameState.PlayerCash[performer.activePlayer] {
		return invalidMoveErr("invalid build: cost %d exceeds player's funds: %d",
			totalCost, gameState.PlayerCash[performer.activePlayer])
	}
	gameState.PlayerCash[performer.activePlayer] -= totalCost
	handler.Log("%s paid a total of $%d for track placements.", handler.ActivePlayerNick(), totalCost)

	// Verify we have not exceeded any component limits by this build
	mapState := newMapState(handler.gameMap, gameState)
	err = checkTownMarkerLimit(mapState)
	if err != nil {
		return err
	}
	err = checkTrackTileLimit(mapState)
	if err != nil {
		return err
	}
	err = handler.checkRouteConnections()
	if err != nil {
		return err
	}

	// Remove ownership of any incomplete links not extended
	for _, link := range gameState.Links {
		if !link.Complete && link.Owner == handler.activePlayer && !performer.extendedLinks[link] {
			handler.Log("%s lost ownership of an incomplete track that started at hex %s",
				handler.ActivePlayerNick(), renderHexCoordinate(link.SourceHex))
			link.Owner = ""
		}
	}

	err = handler.gameMap.PostBuildActionHook(handler.gameState, handler.activePlayer)
	if err != nil {
		return err
	}

	return nil
}
