package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/JackOfMostTrades/eot/backend/tiles"
)

func (performer *buildActionPerformer) attemptTownPlacement(townPlacement *TownPlacement) error {
	hex := townPlacement.Hex
	direction := townPlacement.Track
	mapState := newMapState(performer.gameMap, performer.gameState)
	ts := mapState.GetTileState(hex)

	if !ts.isTown {
		return invalidMoveErr("cannot build town track on a non-town hex")
	}
	if len(ts.routes)+1 > 4 {
		return invalidMoveErr("cannot build more than four tracks on a town hex")
	}
	// Verify that none of the new routes overlap with existing routes
	for _, existingRoute := range ts.routes {
		if existingRoute.Right == townPlacement.Track {
			return invalidMoveErr("cannot build over existing track")
		}
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

	return nil
}

func (performer *buildActionPerformer) attemptTeleportLinkPlacement(teleportLinkPlacement *TeleportLinkPlacement) error {
	hex := teleportLinkPlacement.Hex
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

func (performer *buildActionPerformer) determineTownBuildCost(hex common.Coordinate, tracks []common.Direction) (int, error) {
	ts := newMapState(performer.gameMap, performer.gameState).GetTileState(hex)

	var cost int
	cost = performer.gameMap.GetTownBuildCost(performer.gameState, performer.activePlayer, hex, len(tracks), len(ts.routes) != 0)
	return cost, nil
}

func (performer *buildActionPerformer) determineTrackBuildCost(hex common.Coordinate, tile tiles.TrackTile) (int, error) {
	ts := newMapState(performer.gameMap, performer.gameState).GetTileState(hex)

	hexType := performer.gameMap.GetHexType(hex)
	cost, err := performer.gameMap.GetTrackBuildCost(performer.gameState, performer.activePlayer,
		hexType, hex, tiles.GetTrackType(tile), len(ts.routes) != 0)
	if err != nil {
		return 0, fmt.Errorf("failed to determine cost for placing track tile: %v", err)
	}

	return cost, nil
}

func (performer *buildActionPerformer) attemptTrackPlacement(trackPlacement *TrackPlacement) error {
	tilePlacer := newTilePlacer(performer)
	err := tilePlacer.applyTrackTilePlacement(trackPlacement.Hex,
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

func (handler *confirmMoveHandler) performBuildAction(buildAction *BuildAction) error {

	gameState := handler.gameState
	performer := newBuildActionPerformer(handler.gameMap, handler.gameState, handler.activePlayer)

	// First handle urbanization
	if buildAction.Urbanization != nil {
		if gameState.PlayerActions[handler.activePlayer] != common.URBANIZATION_SPECIAL_ACTION {
			return invalidMoveErr("cannot urbanize without special action")
		}
		if buildAction.Urbanization.City < 0 || buildAction.Urbanization.City >= 8 {
			return invalidMoveErr("invalid city: %d", buildAction.Urbanization.City)
		}

		for _, existingUrb := range gameState.Urbanizations {
			if existingUrb.Hex == buildAction.Urbanization.Hex {
				return invalidMoveErr("cannot urbanize on top of existing urbanization")
			}
			if existingUrb.City == buildAction.Urbanization.City {
				return invalidMoveErr("requested city has already been urbanized")
			}
		}
		if handler.gameMap.GetHexType(buildAction.Urbanization.Hex) != maps.TOWN_HEX_TYPE {
			return invalidMoveErr("must urbanize on town hex")
		}

		gameState.Urbanizations = append(gameState.Urbanizations, buildAction.Urbanization)
		handler.Log("%s urbanizes new city %c at %s",
			handler.ActivePlayerNick(), 'A'+buildAction.Urbanization.City, renderHexCoordinate(buildAction.Urbanization.Hex))

		// Check if there is adjacent incomplete link that becomes completed by this build
		mapState := newMapState(performer.gameMap, performer.gameState)
		for _, direction := range common.ALL_DIRECTIONS {
			adjacentHex := applyDirection(buildAction.Urbanization.Hex, direction)
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
			if link.SourceHex.X == buildAction.Urbanization.Hex.X && link.SourceHex.Y == buildAction.Urbanization.Hex.Y &&
				!link.Complete && len(link.Steps) == 1 {
				gameState.Links = DeleteFromSliceUnordered(i, gameState.Links)
				i -= 1
			}
		}
	}

	// Consolidate placements by hex to determine cost and validity
	townPlacements := make(map[common.Coordinate][]common.Direction)
	for _, townPlacement := range buildAction.TownPlacements {
		townPlacements[townPlacement.Hex] = append(townPlacements[townPlacement.Hex], townPlacement.Track)
	}
	trackPlacements := make(map[common.Coordinate]*TrackPlacement)
	for _, trackPlacement := range buildAction.TrackPlacements {
		if _, ok := trackPlacements[trackPlacement.Hex]; ok {
			return invalidMoveErr("cannot place two tiles on the same hex in a single build phase")
		}
		trackPlacements[trackPlacement.Hex] = trackPlacement
	}

	// Check the number of placements is valid
	placementLimit, err := handler.gameMap.GetBuildLimit(gameState, handler.activePlayer)
	if err != nil {
		return err
	}
	if len(townPlacements)+len(trackPlacements)+len(buildAction.TeleportLinkPlacements) > placementLimit {
		return invalidMoveErr("cannot exceed track placement limit (%d)", placementLimit)
	}

	// Now apply cost
	teleportCosts := make([]int, 0, len(buildAction.TeleportLinkPlacements))
	for _, teleportLinkPlacement := range buildAction.TeleportLinkPlacements {
		cost := handler.gameMap.GetTeleportLinkBuildCost(gameState, handler.activePlayer,
			teleportLinkPlacement.Hex, teleportLinkPlacement.Track)
		if cost == 0 {
			return invalidMoveErr("invalid teleport link placement (no teleport link exists in the target hex/direction)")
		}

		teleportCosts = append(teleportCosts, cost)
	}
	townCosts := make([]int, 0, len(townPlacements))
	for hex, tracks := range townPlacements {
		cost, err := performer.determineTownBuildCost(hex, tracks)
		if err != nil {
			return err
		}
		townCosts = append(townCosts, cost)
	}
	trackCosts := make([]int, 0, len(trackPlacements))
	for hex, tracks := range trackPlacements {
		cost, err := performer.determineTrackBuildCost(hex, tracks.Tile)
		if err != nil {
			return err
		}
		trackCosts = append(trackCosts, cost)
	}
	totalCost := handler.gameMap.GetTotalBuildCost(gameState, handler.activePlayer,
		townCosts, trackCosts, teleportCosts)
	if totalCost > gameState.PlayerCash[performer.activePlayer] {
		return invalidMoveErr("invalid build: cost %d exceeds player's funds: %d",
			totalCost, gameState.PlayerCash[performer.activePlayer])
	}
	gameState.PlayerCash[performer.activePlayer] -= totalCost

	for _, townPlacement := range buildAction.TownPlacements {
		err := performer.attemptTownPlacement(townPlacement)
		if err != nil {
			return err
		}
		handler.Log("%s added track on town hex %s",
			handler.ActivePlayerNick(), renderHexCoordinate(townPlacement.Hex))
	}
	for _, trackPlacement := range buildAction.TrackPlacements {
		err := performer.attemptTrackPlacement(trackPlacement)
		if err != nil {
			return err
		}
		handler.Log("%s added track on hex %s",
			handler.ActivePlayerNick(), renderHexCoordinate(trackPlacement.Hex))
	}
	for _, teleportLinkPlacement := range buildAction.TeleportLinkPlacements {
		err := performer.attemptTeleportLinkPlacement(teleportLinkPlacement)
		if err != nil {
			return err
		}
		handler.Log("%s added teleport link on hex %s",
			handler.ActivePlayerNick(), renderHexCoordinate(teleportLinkPlacement.Hex))
	}

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
