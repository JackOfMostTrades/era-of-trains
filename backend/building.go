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
	ts := performer.mapState[townPlacement.Hex.Y][townPlacement.Hex.X]

	if !ts.HasTown || len(ts.Routes)+len(townPlacement.Tracks) > 4 {
		return ErrInvalidPlacement
	}
	// Verify that none of the new routes overlap with existing routes
	for _, direction := range townPlacement.Tracks {
		for _, existingRoute := range ts.Routes {
			if existingRoute.Right == direction {
				return ErrInvalidPlacement
			}
		}
	}

	var cost int
	if len(ts.Routes) == 0 {
		cost = 1 + len(townPlacement.Tracks)
	} else {
		cost = 3
	}
	if cost > performer.gameState.PlayerCash[performer.gameState.ActivePlayer] {
		return ErrInvalidPlacement
	}
	performer.gameState.PlayerCash[performer.gameState.ActivePlayer] -= cost

	// FIXME: Check component limits (8 town marker tokens)

	for _, direction := range townPlacement.Tracks {
		// If it hits a stop add a new link for the player
		// If it hits player existing track, add this new step to the link and mark it as complete
		// If it hits nothing, add an incomplete new link for the player

		hex := applyDirection(townPlacement.Hex, direction)
		next := performer.mapState[hex.Y][hex.X]
		var link *Link
		if next.IsCity {
			link = &Link{
				SourceHex: townPlacement.Hex,
				Owner:     performer.gameState.ActivePlayer,
				Steps:     []Direction{direction},
				Complete:  true,
			}
			performer.gameState.Links = append(performer.gameState.Links, link)
		} else {
			isJoiningRoute := false
			for _, route := range next.Routes {
				if route.Left == direction.opposite() || route.Right == direction.opposite() {
					// Check that we are not joining into a different player's track
					if route.Link.Owner != "" && route.Link.Owner != performer.gameState.ActivePlayer {
						return ErrInvalidPlacement
					}

					link = route.Link
					route.Link.Steps = append(route.Link.Steps, direction.opposite())
					route.Link.Complete = true
					route.Link.Owner = performer.gameState.ActivePlayer
					isJoiningRoute = true
					break
				}
			}
			if !isJoiningRoute {
				link = &Link{
					SourceHex: townPlacement.Hex,
					Owner:     performer.gameState.ActivePlayer,
					Steps:     []Direction{direction},
					Complete:  false,
				}
				performer.gameState.Links = append(performer.gameState.Links, link)
				performer.extendedLinks[link] = true
			}
		}

		ts.Routes = append(ts.Routes, Route{
			Right: direction,
			Link:  link,
		})
	}

	return nil
}

func (performer *buildActionPerformer) attemptTrackPlacement(trackPlacement *TrackPlacement) error {
	ts := performer.mapState[trackPlacement.Hex.Y][trackPlacement.Hex.X]

	if ts.HasTown || len(ts.Routes)+len(trackPlacement.Tracks) > 2 {
		return ErrInvalidPlacement
	}
	// Verify that none of the new routes overlap with existing routes
	for _, newRoute := range trackPlacement.Tracks {
		for _, existingRoute := range ts.Routes {
			if existingRoute.Left == newRoute[0] || existingRoute.Left == newRoute[1] || existingRoute.Right == newRoute[0] || existingRoute.Right == newRoute[1] {
				return ErrInvalidPlacement
			}
		}
	}

	var cost int
	if len(ts.Routes) == 0 {
		hexType := performer.theMap.Hexes[trackPlacement.Hex.Y][trackPlacement.Hex.X]
		// FIXME: These costs need to be adjusted for direct builds of complex track
		if hexType == PLAINS_HEX_TYPE {
			cost = 2
		} else if hexType == RIVER_HEX_TYPE {
			cost = 3
		} else if hexType == MOUNTAIN_HEX_TYPE {
			cost = 4
		} else {
			return ErrInvalidPlacement
		}
	} else {
		// FIXME: Figure out crossing vs coexisting cost
		cost = 3
	}

	if cost > performer.gameState.PlayerCash[performer.gameState.ActivePlayer] {
		return ErrInvalidPlacement
	}
	performer.gameState.PlayerCash[performer.gameState.ActivePlayer] -= cost

	// FIXME: Check component limits

	for _, newRoute := range trackPlacement.Tracks {
		//        On each side, determine if it hits a stop or extends the player's track
		//          If nothing on either side, this is invalid build
		//          If stop on one side, add this as a new incomplete link
		//          If track on one side, extend existing incomplete link
		//          If track on both sides, join the two incomplete links as a single complete link
		//          If stop on both sides, add new complete track
		//          If track and stop, extend existing incomplete track as a completed link

		leftHex := applyDirection(trackPlacement.Hex, newRoute[0])
		rightHex := applyDirection(trackPlacement.Hex, newRoute[1])
		leftTileState := performer.mapState[leftHex.Y][leftHex.X]
		rightTileState := performer.mapState[rightHex.Y][rightHex.X]

		var link *Link
		if leftTileState.IsCity {
			link = &Link{
				SourceHex: leftHex,
				Owner:     performer.gameState.ActivePlayer,
				Steps:     []Direction{newRoute[0].opposite(), newRoute[1]},
			}
			performer.gameState.Links = append(performer.gameState.Links, link)
			performer.extendedLinks[link] = true
		} else {
			for _, existingRoute := range leftTileState.Routes {
				if existingRoute.Left.opposite() == newRoute[0] || existingRoute.Right.opposite() == newRoute[0] {
					if existingRoute.Link.Owner != "" && existingRoute.Link.Owner != performer.gameState.ActivePlayer {
						return ErrInvalidPlacement
					}
					link = existingRoute.Link
					link.Owner = performer.gameState.ActivePlayer
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
					Owner:     performer.gameState.ActivePlayer,
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
					if existingRoute.Link.Owner != "" && existingRoute.Link.Owner != performer.gameState.ActivePlayer {
						return ErrInvalidPlacement
					}
					if link == nil {
						link = existingRoute.Link
						link.Owner = performer.gameState.ActivePlayer
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
	}

	return nil
}

type buildActionPerformer struct {
	extendedLinks map[*Link]bool
	gameState     *GameState
	theMap        *BasicMap
	mapState      [][]*TileState
}

func newBuildActionPerformer(theMap *BasicMap, gameState *GameState) *buildActionPerformer {

	performer := &buildActionPerformer{
		extendedLinks: make(map[*Link]bool),
		gameState:     gameState,
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

	performer := newBuildActionPerformer(handler.theMap, handler.gameState)
	gameState := handler.gameState

	// First handle urbanization
	if buildAction.Urbanization != nil {
		if gameState.PlayerActions[gameState.ActivePlayer] != URBANIZATION_SPECIAL_ACTION {
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

	startingCash := gameState.PlayerCash[gameState.ActivePlayer]
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

	handler.Log("%s paid a total of $%d for track placements.", handler.ActivePlayerNick(),
		startingCash-gameState.PlayerCash[gameState.ActivePlayer])

	// Remove ownership of any incomplete links not extended
	for _, link := range gameState.Links {
		if !link.Complete && link.Owner == gameState.ActivePlayer && !performer.extendedLinks[link] {
			handler.Log("%s lost ownership of an incomplete track that started at hex (%d,%d)",
				handler.ActivePlayerNick(), link.SourceHex.X, link.SourceHex.Y)
			link.Owner = ""
		}
	}

	return nil
}
