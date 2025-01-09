package maps

import (
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
)

type LogFun = func(format string, a ...any)

type GameMap interface {
	GetWidth() int
	GetHeight() int
	GetHexType(hex common.Coordinate) HexType
	GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color
	GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate
	PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error
	GetTurnLimit(playerCount int) int
	GetSharesLimit() int
	GetGoodsGrowthDiceCount(playerCount int) int
	GetTeleportLink(gameState *common.GameState, src common.Coordinate, direction common.Direction) (*common.Coordinate, common.Direction)

	GetBuildLimit(gameState *common.GameState, player string) (int, error)
	GetTownBuildCost(gameState *common.GameState, player string, hex common.Coordinate, routeCount int, isUpgrade bool) int
	GetTrackBuildCost(gameState *common.GameState, player string, hexType HexType, hex common.Coordinate, trackType common.TrackType, isUpgrade bool) (int, error)
	GetTotalBuildCost(gameState *common.GameState, player string, redirectCosts []int, townCosts []int, trackCosts []int, teleportCosts []int) int
	GetTeleportLinkBuildCost(gameState *common.GameState, player string, hex common.Coordinate, direction common.Direction) int
	GetIncomeReduction(gameState *common.GameState, player string) (int, error)
	PostSetupHook(gameState *common.GameState, randProvider common.RandProvider) error
	PreAuctionHook(gameState *common.GameState, log LogFun) error
	PostBuildActionHook(gameState *common.GameState, player string) error
	PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error
	GetDeliveryBonus(color common.Color) int
	LocationBlocksCubePassage(cube common.Color, hex common.Coordinate) bool
	LocationCanAcceptCube(cube common.Color, hex common.Coordinate) bool
}

type AbstractGameMapImpl struct{}

var _ GameMap = (*AbstractGameMapImpl)(nil)

func (*AbstractGameMapImpl) GetWidth() int {
	return 0
}

func (*AbstractGameMapImpl) GetHeight() int {
	return 0
}

func (*AbstractGameMapImpl) GetHexType(hex common.Coordinate) HexType {
	return OFFBOARD_HEX_TYPE
}

func (*AbstractGameMapImpl) GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color {
	return common.NONE_COLOR
}

func (*AbstractGameMapImpl) GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate {
	return common.Coordinate{X: -1, Y: -1}
}

func (*AbstractGameMapImpl) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	return nil
}

func (*AbstractGameMapImpl) GetTurnLimit(playerCount int) int {
	if playerCount == 6 {
		return 6
	} else if playerCount == 5 {
		return 7
	} else if playerCount == 4 {
		return 8
	}
	return 10
}

func (*AbstractGameMapImpl) GetSharesLimit() int {
	return 15
}

func (*AbstractGameMapImpl) GetGoodsGrowthDiceCount(playerCount int) int {
	return playerCount
}

func (*AbstractGameMapImpl) GetTeleportLink(gameState *common.GameState, src common.Coordinate, direction common.Direction) (*common.Coordinate, common.Direction) {
	return nil, 0
}

func (*AbstractGameMapImpl) GetBuildLimit(gameState *common.GameState, player string) (int, error) {
	action := gameState.PlayerActions[player]
	if action == common.ENGINEER_SPECIAL_ACTION {
		return 4, nil
	}
	return 3, nil
}

func (*AbstractGameMapImpl) GetTownBuildCost(gameState *common.GameState, player string, hex common.Coordinate, routeCount int, isUpgrade bool) int {
	if isUpgrade {
		return 3
	}
	return 1 + routeCount
}

func (*AbstractGameMapImpl) GetTrackBuildCost(gameState *common.GameState, player string, hexType HexType, hex common.Coordinate, trackType common.TrackType, isUpgrade bool) (int, error) {
	if isUpgrade {
		switch trackType {
		case common.SIMPLE_TRACK_TYPE:
			break
		case common.COMPLEX_COEXISTING_TRACK_TYPE:
			return 2, nil
		case common.COMPLEX_CROSSING_TRACK_TYPE:
			return 3, nil
		}
	} else {
		if hexType == PLAINS_HEX_TYPE {
			switch trackType {
			case common.SIMPLE_TRACK_TYPE:
				return 2, nil
			case common.COMPLEX_COEXISTING_TRACK_TYPE:
				return 3, nil
			case common.COMPLEX_CROSSING_TRACK_TYPE:
				return 4, nil
			}
		} else if hexType == RIVER_HEX_TYPE || hexType == HILLS_HEX_TYPE {
			switch trackType {
			case common.SIMPLE_TRACK_TYPE:
				return 3, nil
			case common.COMPLEX_COEXISTING_TRACK_TYPE:
				return 4, nil
			case common.COMPLEX_CROSSING_TRACK_TYPE:
				return 5, nil
			}
		} else if hexType == MOUNTAIN_HEX_TYPE {
			switch trackType {
			case common.SIMPLE_TRACK_TYPE:
				return 4, nil
			case common.COMPLEX_COEXISTING_TRACK_TYPE:
				return 5, nil
			case common.COMPLEX_CROSSING_TRACK_TYPE:
				return 6, nil
			}
		}
	}
	return 0, fmt.Errorf("cannot place track type %v on hex type: %v", trackType, hexType)
}

func (*AbstractGameMapImpl) GetTotalBuildCost(gameState *common.GameState, player string,
	redirectCosts []int, townCosts []int, trackCosts []int, teleportCosts []int) int {

	totalCost := 0
	for _, cost := range redirectCosts {
		totalCost += cost
	}
	for _, cost := range townCosts {
		totalCost += cost
	}
	for _, cost := range trackCosts {
		totalCost += cost
	}
	for _, cost := range teleportCosts {
		totalCost += cost
	}
	return totalCost
}

func (*AbstractGameMapImpl) GetIncomeReduction(gameState *common.GameState, player string) (int, error) {
	income := gameState.PlayerIncome[player]
	var reduction int
	if income <= 10 {
		reduction = 0
	} else if income <= 20 {
		reduction = 2
	} else if income <= 30 {
		reduction = 4
	} else if income <= 40 {
		reduction = 6
	} else if income <= 49 {
		// Yes, this is meant to be 49 and not 50; the ledge for 10 income reduction is 49, per the latest rulebook.
		reduction = 8
	} else {
		reduction = 10
	}
	return reduction, nil
}

func (*AbstractGameMapImpl) GetTeleportLinkBuildCost(gameState *common.GameState, player string, hex common.Coordinate, direction common.Direction) int {
	return 0
}

func (*AbstractGameMapImpl) PostSetupHook(gameState *common.GameState, randProvider common.RandProvider) error {
	return nil
}

func (*AbstractGameMapImpl) PreAuctionHook(gameState *common.GameState, log LogFun) error {
	return nil
}

func (*AbstractGameMapImpl) PostBuildActionHook(gameState *common.GameState, player string) error {
	return nil
}

func (*AbstractGameMapImpl) PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error {
	return nil
}

func (*AbstractGameMapImpl) GetDeliveryBonus(color common.Color) int {
	return 0
}

func (*AbstractGameMapImpl) LocationBlocksCubePassage(cube common.Color, hex common.Coordinate) bool {
	return false
}

func (*AbstractGameMapImpl) LocationCanAcceptCube(cube common.Color, hex common.Coordinate) bool {
	return false
}

func LoadMaps() (map[string]GameMap, error) {
	maps := make(map[string]GameMap)

	rustBelt, err := loadBasicMap("maps/rust_belt.json")
	if err != nil {
		return nil, err
	}
	maps["rust_belt"] = rustBelt

	southernUs, err := loadSouthernUsMap()
	if err != nil {
		return nil, err
	}
	maps["southern_us"] = southernUs

	germany, err := loadGermanyMap()
	if err != nil {
		return nil, err
	}
	maps["germany"] = germany

	scotland, err := loadScotlandMap()
	if err != nil {
		return nil, err
	}
	maps["scotland"] = scotland

	return maps, nil
}
