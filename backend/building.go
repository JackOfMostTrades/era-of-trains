package main

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/JackOfMostTrades/eot/backend/tiles"
)

type Route struct {
	// Default value and meaningless if there is a town. Otherwise, one edge of the route
	Left common.Direction
	// Other edge of the route
	Right common.Direction
	// What link is this a part of?
	Link *common.Link
}

type TileState struct {
	IsCity  bool
	HasTown bool
	Routes  []Route
}

func (performer *buildActionPerformer) attemptTownPlacement(townPlacement *TownPlacement) error {
	hex := townPlacement.Hex
	direction := townPlacement.Track
	ts := performer.mapState[hex.Y][hex.X]

	if !ts.HasTown {
		return invalidMoveErr("cannot build town track on a non-town hex")
	}
	if len(ts.Routes)+1 > 4 {
		return invalidMoveErr("cannot build more than four tracks on a town hex")
	}
	// Verify that none of the new routes overlap with existing routes
	for _, existingRoute := range ts.Routes {
		if existingRoute.Right == townPlacement.Track {
			return invalidMoveErr("cannot build over existing track")
		}
	}

	// If it hits a stop add a new link for the player
	// If it hits player existing track, then mark it as complete
	// If it hits nothing, add an incomplete new link for the player

	nextHex := applyDirection(hex, direction)
	next := performer.mapState[nextHex.Y][nextHex.X]
	var link *common.Link
	if next.IsCity {
		link = &common.Link{
			SourceHex: hex,
			Owner:     performer.activePlayer,
			Steps:     []common.Direction{direction},
			Complete:  true,
		}
		performer.gameState.Links = append(performer.gameState.Links, link)
	} else {
		isJoiningRoute := false
		for _, route := range next.Routes {
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

	ts.Routes = append(ts.Routes, Route{
		Left:  direction,
		Right: direction,
		Link:  link,
	})

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
	ts := performer.mapState[hex.Y][hex.X]

	var cost int
	cost = performer.gameMap.GetTownBuildCost(performer.gameState, performer.activePlayer, hex, len(tracks), len(ts.Routes) != 0)
	return cost, nil
}

func (performer *buildActionPerformer) determineTrackBuildCost(hex common.Coordinate, tile tiles.TrackTile) (int, error) {
	ts := performer.mapState[hex.Y][hex.X]

	hexType := performer.gameMap.GetHexType(hex)
	cost, err := performer.gameMap.GetTrackBuildCost(performer.gameState, performer.activePlayer,
		hexType, hex, tiles.GetTrackType(tile), len(ts.Routes) != 0)
	if err != nil {
		return 0, fmt.Errorf("failed to determine cost for placing track tile: %v", err)
	}

	return cost, nil
}

func (performer *buildActionPerformer) attemptTrackPlacement(trackPlacement *TrackPlacement) error {
	tilePlacer, err := newTilePlacer(performer)
	if err != nil {
		return err
	}
	err = tilePlacer.applyTrackTilePlacement(trackPlacement.Hex,
		trackPlacement.Tile,
		trackPlacement.Rotation,
		performer.activePlayer,
		performer.gameState)
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
	mapState      [][]*TileState
}

func newBuildActionPerformer(gameMap maps.GameMap, gameState *common.GameState, activePlayer string) *buildActionPerformer {

	performer := &buildActionPerformer{
		extendedLinks: make(map[*common.Link]bool),
		gameState:     gameState,
		activePlayer:  activePlayer,
		gameMap:       gameMap,
		mapState:      make([][]*TileState, gameMap.GetHeight()),
	}

	for y := range performer.mapState {
		performer.mapState[y] = make([]*TileState, gameMap.GetWidth())
		for x := range performer.mapState[y] {
			hexType := gameMap.GetHexType(common.Coordinate{X: x, Y: y})
			performer.mapState[y][x] = &TileState{
				IsCity:  hexType == maps.CITY_HEX_TYPE,
				HasTown: hexType == maps.TOWN_HEX_TYPE,
				Routes:  nil,
			}
		}
	}
	for _, urb := range gameState.Urbanizations {
		performer.mapState[urb.Hex.Y][urb.Hex.X].IsCity = true
	}
	for _, link := range gameState.Links {
		hex := link.SourceHex
		tileState := performer.mapState[hex.Y][hex.X]
		if tileState.HasTown {
			tileState.Routes = append(tileState.Routes, Route{
				Left:  link.Steps[0],
				Right: link.Steps[0],
				Link:  link,
			})
		}
		for idx := 1; idx < len(link.Steps); idx++ {
			hex = applyDirection(hex, link.Steps[idx-1])
			tileState = performer.mapState[hex.Y][hex.X]
			if tileState.IsCity {
				// Do nothing
			} else if tileState.HasTown {
				tileState.Routes = append(tileState.Routes, Route{
					Left:  link.Steps[idx-1].Opposite(),
					Right: link.Steps[idx-1].Opposite(),
					Link:  link,
				})
			} else {
				// Ordinary track in this tile
				tileState.Routes = append(tileState.Routes, Route{
					Left:  link.Steps[idx-1].Opposite(),
					Right: link.Steps[idx],
					Link:  link,
				})
			}
		}
		// If this is a complete link that ends at a town hex, also add the last step to the town hex
		if link.Complete {
			lastStep := link.Steps[len(link.Steps)-1]
			hex = applyDirection(hex, lastStep)
			tileState = performer.mapState[hex.Y][hex.X]
			if tileState.HasTown && !tileState.IsCity {
				tileState.Routes = append(tileState.Routes, Route{
					Left:  lastStep.Opposite(),
					Right: lastStep.Opposite(),
					Link:  link,
				})
			}
		}

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
		ts := performer.mapState[buildAction.Urbanization.Hex.Y][buildAction.Urbanization.Hex.X]
		ts.IsCity = true
		ts.Routes = nil
		handler.Log("%s urbanizes new city %c at %s",
			handler.ActivePlayerNick(), 'A'+buildAction.Urbanization.City, renderHexCoordinate(buildAction.Urbanization.Hex))

		// Check if there is adjacent incomplete link that becomes completed by this build
		for _, direction := range common.ALL_DIRECTIONS {
			adjacentHex := applyDirection(buildAction.Urbanization.Hex, direction)
			if adjacentHex.Y >= 0 && adjacentHex.Y < len(performer.mapState) &&
				adjacentHex.X >= 0 && adjacentHex.X < len(performer.mapState[adjacentHex.Y]) {
				ts := performer.mapState[adjacentHex.Y][adjacentHex.X]
				for _, route := range ts.Routes {
					if route.Left == direction.Opposite() || route.Right == direction.Opposite() {
						route.Link.Complete = true
					}
				}
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

	// For each placement...
	//   If it is a town...
	//     (Verify upgrade is valid and): determine new player's routes; for each:
	//        If it hits a stop add a new link for the player
	//        If it hits player exist track, add this new step to the link and mark it as complete
	//        If it hits nothing, add an incomplete new link for the player
	//   Else (not a stop)
	//     (Verify upgrade is valid and): determine new player's routes; for each:
	//        On each side, determine if it hits a stop or extends the player's track
	//          If nothing on either side, this is invalid build
	//          If stop on one side, add this as a new incomplete link
	//          If track on one side, extend existing incomplete link
	//          If track on both sides, join the two incomplete links as a single complete link
	//          If stop on both sides, add new complete track
	//          If track and stop, extend existing incomplete track as a completed link

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
	err = checkTownMarkerLimit(performer.mapState)
	if err != nil {
		return err
	}
	err = checkTrackTileLimit(performer.mapState)
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
