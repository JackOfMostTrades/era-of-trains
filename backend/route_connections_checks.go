package main

func (handler *confirmMoveHandler) checkRouteConnections(mapState [][]TileState) error {
	for _, link := range handler.gameState.Links {
		if link.Owner != handler.activePlayer {
			continue
		}

		startHex := link.SourceHex
		if mapState[startHex.Y][startHex.X].IsCity {
			continue
		}
		endHex := startHex
		for _, step := range link.Steps {
			endHex = applyMapDirection(handler.gameMap, handler.gameState, endHex, step)
		}
		if mapState[endHex.Y][endHex.X].IsCity {
			continue
		}
	}

	return nil
}
