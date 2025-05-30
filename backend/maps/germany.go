package maps

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
)

type germanyMap struct {
	*basicMap
}

func (m *germanyMap) PostSetupHook(gameState *common.GameState, randProvider common.RandProvider) error {
	portColors := make([]common.Color, 6)
	for i := 0; i < len(portColors); i++ {
		cube, err := gameState.DrawCube(randProvider)
		if err != nil {
			return err
		}
		portColors[i] = cube
	}
	if gameState.MapState == nil {
		gameState.MapState = make(map[string]interface{})
	}
	gameState.MapState["portColors"] = portColors
	return nil
}

func (m *germanyMap) getPortNumber(hex common.Coordinate) int {
	mapData := m.Hexes[hex.Y][hex.X].MapData
	if mapData != nil {
		if portNumber, ok := mapData["portNumber"]; ok {
			if portNumberVal, ok := portNumber.(float64); ok {
				return int(portNumberVal)
			}
		}
	}

	return 0
}

func (m *germanyMap) GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color {
	port := m.getPortNumber(hex)
	if port == 0 {
		return m.basicMap.GetCityColorForHex(gameState, hex)
	}
	if gameState.MapState != nil {
		if portColors, ok := gameState.MapState["portColors"]; ok {
			if portColorsVal, ok := portColors.([]interface{}); ok {
				val := portColorsVal[port-1]
				if colAsFloat, ok := val.(float64); ok {
					return common.Color(int(colAsFloat))
				}
			}
		}
	}

	return common.NONE_COLOR
}

func (m *germanyMap) LocationBlocksCubePassage(cube common.Color, hex common.Coordinate) bool {
	port := m.getPortNumber(hex)
	if port != 0 {
		return true
	}
	return false
}

func (m *germanyMap) GetBuildLimit(gameState *common.GameState, player string) (int, error) {
	return 3, nil
}

func (m *germanyMap) GetTotalBuildCost(gameState *common.GameState, player string, costs []int) int {
	if gameState.PlayerActions[player] == common.ENGINEER_SPECIAL_ACTION {
		maxCost := 0
		totalCost := 0
		for _, cost := range costs {
			totalCost += cost
			if cost > maxCost {
				maxCost = cost
			}
		}
		totalCost -= maxCost / 2
		return totalCost
	} else {
		return m.basicMap.GetTotalBuildCost(gameState, player, costs)
	}
}

func (b *germanyMap) PostBuildActionHook(gameState *common.GameState, player string) error {
	for _, link := range gameState.Links {
		if !link.Complete {
			return fmt.Errorf("all links must be complete on this map")
		}
	}
	return nil
}

func (b *germanyMap) PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error {
	err := b.basicMap.PostGoodsGrowthHook(gameState, randProvider, log)
	if err != nil {
		return err
	}

	color, err := gameState.DrawCube(randProvider)
	if err != nil {
		return err
	}
	if color == common.NONE_COLOR {
		log("No cube was drawn to add to Berlin because the bag is empty.", color.String())
	} else {
		gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
			Color: color,
			Hex:   common.Coordinate{X: 4, Y: 9},
		})
		log("A %s cube was drawn and added to Berlin.", color.String())
	}

	return nil
}

func loadGermanyMap() (GameMap, error) {
	b, err := loadBasicMap("maps/germany.json")
	if err != nil {
		return nil, err
	}
	return &germanyMap{b}, nil
}
