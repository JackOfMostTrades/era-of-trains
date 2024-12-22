package main

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
)

type Route struct {
	// Default value and meaningless if there is a town. Otherwise, one edge of the route
	Left Direction
	// Other edge of the route
	Right Direction
	// What link is this a part of?
	Link *Link
}

type TileState struct {
	IsCity  bool
	HasTown bool
	Routes  []Route
}

type PlacementResult struct {
	ValidPlacement bool
	NewRoutes      []Route
	Cost           int
	NewLinks       []*Link
	LinksToDelete  []*Link
}

var ErrInvalidPlacement = errors.New("invalid placement")

func (performer *buildActionPerformer) attemptTownPlacement(townPlacement *TownPlacement) error {
	hex := townPlacement.Hex
	direction := townPlacement.Track
	ts := performer.mapState[hex.Y][hex.X]

	if !ts.HasTown || len(ts.Routes)+1 > 4 {
		return ErrInvalidPlacement
	}
	// Verify that none of the new routes overlap with existing routes
	for _, existingRoute := range ts.Routes {
		if existingRoute.Right == townPlacement.Track {
			return ErrInvalidPlacement
		}
	}

	// FIXME: Check component limits (8 town marker tokens)

	// If it hits a stop add a new link for the player
	// If it hits player existing track, add this new step to the link and mark it as complete
	// If it hits nothing, add an incomplete new link for the player

	nextHex := applyDirection(hex, direction)
	next := performer.mapState[nextHex.Y][nextHex.X]
	var link *Link
	if next.IsCity {
		link = &Link{
			SourceHex: hex,
			Owner:     performer.activePlayer,
			Steps:     []Direction{direction},
			Complete:  true,
		}
		performer.gameState.Links = append(performer.gameState.Links, link)
	} else {
		isJoiningRoute := false
		for _, route := range next.Routes {
			if route.Left == direction.opposite() || route.Right == direction.opposite() {
				// Check that we are not joining into a different player's track
				if route.Link.Owner != "" && route.Link.Owner != performer.activePlayer {
					return ErrInvalidPlacement
				}

				link = route.Link
				route.Link.Steps = append(route.Link.Steps, direction.opposite())
				route.Link.Complete = true
				route.Link.Owner = performer.activePlayer
				isJoiningRoute = true
				break
			}
		}
		if !isJoiningRoute {
			link = &Link{
				SourceHex: hex,
				Owner:     performer.activePlayer,
				Steps:     []Direction{direction},
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

/*
[[Direction.NORTH, Direction.SOUTH], [Direction.SOUTH_WEST, Direction.NORTH_EAST]],
    // Gentle X
    [[Direction.NORTH, Direction.SOUTH_EAST], [Direction.NORTH_EAST, Direction.SOUTH]],
    // Bow and arrow
    [[Direction.NORTH, Direction.SOUTH], [Direction.SOUTH_WEST, Direction.SOUTH_EAST]],
*/

type trackTileType int

const (
	SIMPLE_TRACK_TILE_TYPE = iota
	COMPLEX_CROSSING_TILE_TYPE
	COMPLEX_COEXISTING_TILE_TYPE
)

func routesEqual(a [][2]Direction, b [][2]Direction) bool {
	if len(a) != len(b) {
		return false
	}
	for _, trackA := range a {
		found := false
		for _, trackB := range b {
			if (trackA[0] == trackB[0] && trackA[1] == trackB[1]) ||
				(trackA[0] == trackB[1] && trackA[1] == trackB[0]) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func getTileType(routes [][2]Direction) trackTileType {
	if len(routes) < 2 {
		return SIMPLE_TRACK_TILE_TYPE
	}

	complexCrossingTiles := [][][2]Direction{
		// X
		{{NORTH, SOUTH}, {SOUTH_WEST, NORTH_EAST}},
		// Gentle X
		{{NORTH, SOUTH_EAST}, {NORTH_EAST, SOUTH}},
		// Bow and arrow
		{{NORTH, SOUTH}, {SOUTH_WEST, SOUTH_EAST}},
	}
	for _, tile := range complexCrossingTiles {
		for rotation := 0; rotation < 6; rotation++ {
			rotatedRoutes := make([][2]Direction, 0, len(tile))
			for _, route := range tile {
				rotatedRoutes = append(rotatedRoutes, [2]Direction{
					Direction((int(route[0]) + rotation) % 6), Direction((int(route[1]) + rotation) % 6),
				})
			}
			if routesEqual(rotatedRoutes, routes) {
				return COMPLEX_CROSSING_TILE_TYPE
			}
		}
	}

	return COMPLEX_COEXISTING_TILE_TYPE
}

func (performer *buildActionPerformer) determineTownBuildCost(hex Coordinate, tracks []Direction) (int, error) {
	ts := performer.mapState[hex.Y][hex.X]

	var cost int
	if len(ts.Routes) == 0 {
		cost = 1 + len(tracks)
	} else {
		cost = 3
	}
	return cost, nil
}

func (performer *buildActionPerformer) determineTrackBuildCost(hex Coordinate, tracks [][2]Direction) (int, error) {
	ts := performer.mapState[hex.Y][hex.X]

	// Max number of routes on a tile is 2
	if len(ts.Routes)+len(tracks) > 2 {
		return 0, ErrInvalidPlacement
	}

	var cost int
	if len(ts.Routes) == 0 {
		hexType := performer.theMap.Hexes[hex.Y][hex.X]
		tileType := getTileType(tracks)
		if hexType == PLAINS_HEX_TYPE {
			switch tileType {
			case SIMPLE_TRACK_TILE_TYPE:
				cost = 2
			case COMPLEX_COEXISTING_TILE_TYPE:
				cost = 3
			case COMPLEX_CROSSING_TILE_TYPE:
				cost = 4
			}
		} else if hexType == RIVER_HEX_TYPE {
			switch tileType {
			case SIMPLE_TRACK_TILE_TYPE:
				cost = 3
			case COMPLEX_COEXISTING_TILE_TYPE:
				cost = 4
			case COMPLEX_CROSSING_TILE_TYPE:
				cost = 5
			}
		} else if hexType == MOUNTAIN_HEX_TYPE {
			switch tileType {
			case SIMPLE_TRACK_TILE_TYPE:
				cost = 4
			case COMPLEX_COEXISTING_TILE_TYPE:
				cost = 5
			case COMPLEX_CROSSING_TILE_TYPE:
				cost = 6
			}
		} else {
			return 0, ErrInvalidPlacement
		}
	} else {
		allRoutes := make([][2]Direction, 0, len(ts.Routes)+len(tracks))
		for _, route := range ts.Routes {
			allRoutes = append(allRoutes, [2]Direction{route.Left, route.Right})
		}
		for _, track := range tracks {
			allRoutes = append(allRoutes, track)
		}
		tileType := getTileType(allRoutes)
		if tileType == COMPLEX_CROSSING_TILE_TYPE {
			cost = 3
		} else {
			cost = 2
		}
	}
	if cost == 0 {
		return 0, fmt.Errorf("failed to determine cost for placing track tile")
	}

	return cost, nil
}

func (performer *buildActionPerformer) attemptTrackPlacement(trackPlacement *TrackPlacement) error {
	hex := trackPlacement.Hex
	newRoute := trackPlacement.Track
	ts := performer.mapState[hex.Y][hex.X]

	if ts.HasTown || len(ts.Routes)+1 > 2 {
		return ErrInvalidPlacement
	}
	// Verify that none of the new routes overlap with existing routes
	for _, existingRoute := range ts.Routes {
		if existingRoute.Left == newRoute[0] || existingRoute.Left == newRoute[1] || existingRoute.Right == newRoute[0] || existingRoute.Right == newRoute[1] {
			return ErrInvalidPlacement
		}
	}

	// FIXME: Check component limits

	//        On each side, determine if it hits a stop or extends the player's track
	//          If nothing on either side, this is invalid build
	//          If stop on one side, add this as a new incomplete link
	//          If track on one side, extend existing incomplete link
	//          If track on both sides, join the two incomplete links as a single complete link
	//          If stop on both sides, add new complete track
	//          If track and stop, extend existing incomplete track as a completed link

	leftHex := applyDirection(hex, newRoute[0])
	rightHex := applyDirection(hex, newRoute[1])
	leftTileState := performer.mapState[leftHex.Y][leftHex.X]
	rightTileState := performer.mapState[rightHex.Y][rightHex.X]

	var link *Link
	if leftTileState.IsCity {
		link = &Link{
			SourceHex: leftHex,
			Owner:     performer.activePlayer,
			Steps:     []Direction{newRoute[0].opposite(), newRoute[1]},
		}
		performer.gameState.Links = append(performer.gameState.Links, link)
		performer.extendedLinks[link] = true
	} else {
		for _, existingRoute := range leftTileState.Routes {
			if existingRoute.Left.opposite() == newRoute[0] || existingRoute.Right.opposite() == newRoute[0] {
				if existingRoute.Link.Owner != "" && existingRoute.Link.Owner != performer.activePlayer {
					return ErrInvalidPlacement
				}
				link = existingRoute.Link
				link.Owner = performer.activePlayer
				link.Steps = append(link.Steps, newRoute[1])
				performer.extendedLinks[link] = true
				break
			}
		}
	}

	if rightTileState.IsCity {
		if link == nil {
			link = &Link{
				SourceHex: rightHex,
				Owner:     performer.activePlayer,
				Steps:     []Direction{newRoute[1].opposite(), newRoute[0]},
			}
			performer.gameState.Links = append(performer.gameState.Links, link)
			performer.extendedLinks[link] = true
		} else {
			link.Complete = true
		}
	} else {
		for _, existingRoute := range rightTileState.Routes {
			if existingRoute.Left.opposite() == newRoute[1] || existingRoute.Right.opposite() == newRoute[1] {
				if existingRoute.Link.Owner != "" && existingRoute.Link.Owner != performer.activePlayer {
					return ErrInvalidPlacement
				}
				if link == nil {
					link = existingRoute.Link
					link.Owner = performer.activePlayer
					link.Steps = append(link.Steps, newRoute[0])
					performer.extendedLinks[link] = true
					break
				} else {
					// Delete the right-hand link since that will be consumed onto the left-hand link
					performer.gameState.Links = DeleteFromSliceUnordered(
						slices.Index(performer.gameState.Links, existingRoute.Link), performer.gameState.Links)
					// Add the old link to the end of the new link
					for idx := len(existingRoute.Link.Steps) - 2; idx >= 0; idx-- {
						link.Steps = append(link.Steps, existingRoute.Link.Steps[idx].opposite())
					}
					link.Complete = true
				}
			}
		}
		if link == nil {
			// No link from left side, no link from right side: invalid placement
			return ErrInvalidPlacement
		}
	}

	ts.Routes = append(ts.Routes, Route{
		Left:  newRoute[0],
		Right: newRoute[1],
		Link:  link,
	})

	return nil
}

type buildActionPerformer struct {
	extendedLinks map[*Link]bool
	gameState     *GameState
	activePlayer  string
	theMap        *BasicMap
	mapState      [][]*TileState
}

func newBuildActionPerformer(theMap *BasicMap, gameState *GameState, activePlayer string) *buildActionPerformer {

	performer := &buildActionPerformer{
		extendedLinks: make(map[*Link]bool),
		gameState:     gameState,
		activePlayer:  activePlayer,
		theMap:        theMap,
		mapState:      make([][]*TileState, theMap.Height),
	}

	for y := range performer.mapState {
		performer.mapState[y] = make([]*TileState, theMap.Width)
		for x := range performer.mapState[y] {
			performer.mapState[y][x] = &TileState{
				IsCity:  theMap.Hexes[y][x] == CITY_HEX_TYPE,
				HasTown: theMap.Hexes[y][x] == TOWN_HEX_TYPE,
				Routes:  nil,
			}
		}
	}
	for _, city := range theMap.Cities {
		performer.mapState[city.Coordinate.Y][city.Coordinate.X].IsCity = true
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
					Left:  link.Steps[idx-1].opposite(),
					Right: link.Steps[idx-1].opposite(),
					Link:  link,
				})
			} else {
				// Ordinary track in this tile
				tileState.Routes = append(tileState.Routes, Route{
					Left:  link.Steps[idx-1].opposite(),
					Right: link.Steps[idx],
					Link:  link,
				})
			}
		}
	}

	return performer
}

func (handler *confirmMoveHandler) performBuildAction(buildAction *BuildAction) error {

	gameState := handler.gameState

	// First handle urbanization
	if buildAction.Urbanization != nil {
		if gameState.PlayerActions[handler.activePlayer] != URBANIZATION_SPECIAL_ACTION {
			return &HttpError{"cannot urbanize without special action", http.StatusBadRequest}
		}
		if buildAction.Urbanization.City < 0 || buildAction.Urbanization.City >= 8 {
			return &HttpError{fmt.Sprintf("invalid city: %d", buildAction.Urbanization.City), http.StatusBadRequest}
		}

		for _, existingUrb := range gameState.Urbanizations {
			if existingUrb.Hex == buildAction.Urbanization.Hex {
				return &HttpError{"cannot urbanize on top of existing urbanization", http.StatusBadRequest}
			}
			if existingUrb.City == buildAction.Urbanization.City {
				return &HttpError{"requested city has already been urbanized", http.StatusBadRequest}
			}
		}
		if handler.theMap.Hexes[buildAction.Urbanization.Hex.Y][buildAction.Urbanization.Hex.X] != TOWN_HEX_TYPE {
			return &HttpError{"must urbanize on town hex", http.StatusBadRequest}
		}

		gameState.Urbanizations = append(gameState.Urbanizations, buildAction.Urbanization)
		handler.Log("%s urbanizes new city %c on hex (%d,%d)",
			handler.ActivePlayerNick(), 'A'+buildAction.Urbanization.City, buildAction.Urbanization.Hex.X, buildAction.Urbanization.Hex.Y)
	}

	// Consolidate placements by hex to determine cost and validity
	townPlacements := make(map[Coordinate][]Direction)
	for _, townPlacement := range buildAction.TownPlacements {
		townPlacements[townPlacement.Hex] = append(townPlacements[townPlacement.Hex], townPlacement.Track)
	}
	trackPlacements := make(map[Coordinate][][2]Direction)
	for _, trackPlacement := range buildAction.TrackPlacements {
		trackPlacements[trackPlacement.Hex] = append(trackPlacements[trackPlacement.Hex], trackPlacement.Track)
	}

	// Check the number of placements is valid
	placementLimit := 3
	if gameState.PlayerActions[handler.activePlayer] == ENGINEER_SPECIAL_ACTION {
		placementLimit = 4
	}
	if len(townPlacements)+len(trackPlacements) > placementLimit {
		return &HttpError{fmt.Sprintf("cannot exceed track placement limit (%d)", placementLimit), http.StatusBadRequest}
	}

	performer := newBuildActionPerformer(handler.theMap, handler.gameState, handler.activePlayer)

	// Now apply cost
	totalCost := 0
	for hex, tracks := range townPlacements {
		cost, err := performer.determineTownBuildCost(hex, tracks)
		if err != nil {
			return err
		}
		totalCost += cost
	}
	for hex, tracks := range trackPlacements {
		cost, err := performer.determineTrackBuildCost(hex, tracks)
		if err != nil {
			return err
		}
		totalCost += cost
	}
	if totalCost > gameState.PlayerCash[performer.activePlayer] {
		return ErrInvalidPlacement
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
			if errors.Is(err, ErrInvalidPlacement) {
				return &HttpError{"invalid tile placement", http.StatusBadRequest}
			}
			return err
		}
		handler.Log("%s added track on town hex (%d,%d)",
			handler.ActivePlayerNick(), townPlacement.Hex.X, townPlacement.Hex.Y)
	}
	for _, trackPlacement := range buildAction.TrackPlacements {
		err := performer.attemptTrackPlacement(trackPlacement)
		if err != nil {
			if errors.Is(err, ErrInvalidPlacement) {
				return &HttpError{"invalid tile placement", http.StatusBadRequest}
			}
			return err
		}
		handler.Log("%s added track on hex (%d,%d)",
			handler.ActivePlayerNick(), trackPlacement.Hex.X, trackPlacement.Hex.Y)
	}

	handler.Log("%s paid a total of $%d for track placements.", handler.ActivePlayerNick(), totalCost)

	// Remove ownership of any incomplete links not extended
	for _, link := range gameState.Links {
		if !link.Complete && link.Owner == handler.activePlayer && !performer.extendedLinks[link] {
			handler.Log("%s lost ownership of an incomplete track that started at hex (%d,%d)",
				handler.ActivePlayerNick(), link.SourceHex.X, link.SourceHex.Y)
			link.Owner = ""
		}
	}

	return nil
}
