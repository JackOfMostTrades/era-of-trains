package maps

import "github.com/JackOfMostTrades/eot/backend/common"

type LogFun = func(format string, a ...any)

type GameMap interface {
	GetWidth() int
	GetHeight() int
	GetHexType(hex common.Coordinate) HexType
	GetCityColorForHex(hex common.Coordinate) common.Color
	GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate
	PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error

	GetIncomeReduction(gameState *common.GameState, player string) (int, error)
	PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error
	GetDeliveryBonus(color common.Color) int
	CanAcceptCube(cube common.Color, hex common.Coordinate) bool
}

type AbstractGameMapImpl struct{}

func (*AbstractGameMapImpl) GetWidth() int {
	return 0
}

func (*AbstractGameMapImpl) GetHeight() int {
	return 0
}

func (*AbstractGameMapImpl) GetHexType(hex common.Coordinate) HexType {
	return OFFBOARD_HEX_TYPE
}

func (*AbstractGameMapImpl) GetCityColorForHex(hex common.Coordinate) common.Color {
	return common.NONE_COLOR
}

func (*AbstractGameMapImpl) GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate {
	return common.Coordinate{X: -1, Y: -1}
}

func (*AbstractGameMapImpl) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	return nil
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
	} else if income <= 50 {
		reduction = 8
	} else {
		reduction = 10
	}
	return reduction, nil
}

func (*AbstractGameMapImpl) PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error {
	return nil
}

func (*AbstractGameMapImpl) GetDeliveryBonus(color common.Color) int {
	return 0
}

func (*AbstractGameMapImpl) CanAcceptCube(cube common.Color, hex common.Coordinate) bool {
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

	return maps, nil
}
