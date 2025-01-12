package maps

import (
	"github.com/JackOfMostTrades/eot/backend/common"
)

type southernUsMap struct {
	*basicMap
}

func (b *southernUsMap) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	err := b.basicMap.PopulateStartingCubes(gameState, randProvider)
	if err != nil {
		return err
	}

	/*
	 * > Place a white cube (representing Cotton) in every Town.
	 */
	for x := 0; x < b.GetWidth(); x++ {
		for y := 0; y < b.GetHeight(); y++ {
			hex := common.Coordinate{X: x, Y: y}
			if b.GetHexType(hex) == TOWN_HEX_TYPE {
				gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
					Color: common.WHITE,
					Hex:   hex,
				})
			}
		}
	}
	return nil
}

func (b *southernUsMap) GetDeliveryBonus(color common.Color) int {
	if color == common.WHITE {
		return 1
	}
	return 0
}

func (b *southernUsMap) GetIncomeReduction(gameState *common.GameState, player string) (int, error) {
	reduction, err := b.basicMap.GetIncomeReduction(gameState, player)
	if err != nil {
		return 0, err
	}
	if gameState.TurnNumber == 4 {
		reduction *= 2
	}
	return reduction, nil
}

func (b *southernUsMap) PostGoodsGrowthHook(gameState *common.GameState, randProvider common.RandProvider, log LogFun) error {
	err := b.basicMap.PostGoodsGrowthHook(gameState, randProvider, log)
	if err != nil {
		return err
	}

	/*
	 * > On Turns 1-4, Atlanta always receives 1 Goods cube every turn, drawn directly from the bag, in addition to any
	 * > Goods from the Goods display.
	 */
	if gameState.TurnNumber <= 4 {
		color, err := gameState.DrawCube(randProvider)
		if err != nil {
			return err
		}
		if color != common.NONE_COLOR {
			gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
				Color: color,
				Hex:   common.Coordinate{X: 4, Y: 9},
			})
			log("A %s cube was drawn and added to Atlanta.", color.String())
		}
	}

	return nil
}

func (b *southernUsMap) isWhiteCity(hex common.Coordinate) bool {
	if hex.X == 0 && hex.Y == 22 {
		return true
	}
	if hex.X == 2 && hex.Y == 20 {
		return true
	}
	if hex.X == 7 && hex.Y == 11 {
		return true
	}
	if hex.X == 7 && hex.Y == 14 {
		return true
	}
	return false
}

func (b *southernUsMap) LocationBlocksCubePassage(cube common.Color, hex common.Coordinate) bool {
	if cube == common.WHITE && b.isWhiteCity(hex) {
		return true
	}
	return false
}

func (b *southernUsMap) LocationCanAcceptCube(cube common.Color, hex common.Coordinate) bool {
	if cube == common.WHITE && b.isWhiteCity(hex) {
		return true
	}
	return false
}

func loadSouthernUsMap() (GameMap, error) {
	b, err := loadBasicMap("maps/southern_us.json")
	if err != nil {
		return nil, err
	}
	return &southernUsMap{b}, nil
}
