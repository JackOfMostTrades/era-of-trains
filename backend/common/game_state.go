package common

import "fmt"

type GamePhase int

const (
	SHARES_GAME_PHASE GamePhase = iota + 1
	AUCTION_GAME_PHASE
	CHOOSE_SPECIAL_ACTIONS_GAME_PHASE
	BUILDING_GAME_PHASE
	MOVING_GOODS_GAME_PHASE
	GOODS_GROWTH_GAME_PHASE
)

type SpecialAction string

const (
	FIRST_MOVE_SPECIAL_ACTION      SpecialAction = "first_move"
	FIRST_BUILD_SPECIAL_ACTION     SpecialAction = "first_build"
	ENGINEER_SPECIAL_ACTION        SpecialAction = "engineer"
	LOCO_SPECIAL_ACTION            SpecialAction = "loco"
	URBANIZATION_SPECIAL_ACTION    SpecialAction = "urbanization"
	PRODUCTION_SPECIAL_ACTION      SpecialAction = "production"
	TURN_ORDER_PASS_SPECIAL_ACTION SpecialAction = "turn_order_pass"
)

var ALL_SPECIAL_ACTIONS = []SpecialAction{FIRST_MOVE_SPECIAL_ACTION, FIRST_BUILD_SPECIAL_ACTION, ENGINEER_SPECIAL_ACTION, LOCO_SPECIAL_ACTION,
	URBANIZATION_SPECIAL_ACTION, PRODUCTION_SPECIAL_ACTION, TURN_ORDER_PASS_SPECIAL_ACTION}

type Link struct {
	SourceHex Coordinate  `json:"sourceHex"`
	Steps     []Direction `json:"steps"`
	Owner     string      `json:"owner"`
	Complete  bool        `json:"complete"`
}

type Urbanization struct {
	Hex Coordinate `json:"hex"`
	// A=0, B=1, ...
	City int `json:"city"`
}

type BoardCube struct {
	Color Color      `json:"color"`
	Hex   Coordinate `json:"hex"`
}

type GameState struct {
	PlayerColor   map[string]int           `json:"playerColor"`
	PlayerOrder   []string                 `json:"playerOrder"`
	PlayerShares  map[string]int           `json:"playerShares"`
	PlayerLoco    map[string]int           `json:"playerLoco"`
	PlayerIncome  map[string]int           `json:"playerIncome"`
	PlayerActions map[string]SpecialAction `json:"playerActions"`
	PlayerCash    map[string]int           `json:"playerCash"`

	// Map from player ID to their last bid
	AuctionState map[string]int `json:"auctionState"`

	GamePhase  GamePhase `json:"gamePhase"`
	TurnNumber int       `json:"turnNumber"`

	// Which round of moving goods are we in (0 or 1)
	MovingGoodsRound int `json:"movingGoodsRound"`
	// Which users did loco during move goods (to ensure they don't double-loco)
	PlayerHasDoneLoco map[string]bool `json:"playerHasDoneLoco"`

	Links         []*Link         `json:"links"`
	Urbanizations []*Urbanization `json:"urbanizations"`
	// Map from color to number of cubes of that color in the bag
	CubeBag map[Color]int `json:"cubeBag"`
	// Cubes present on the board
	Cubes []*BoardCube `json:"cubes"`
	// Cubes present on the goods-growth chart, 1-6 white, 7-12 black, 13-20 new cities
	GoodsGrowth [][]Color `json:"goodsGrowth"`
	// If cubes have been drawn for the production action, these are the cubes
	ProductionCubes []Color `json:"productionCubes"`

	// Untyped JSON object that maps can use for map-custom state
	MapState map[string]interface{} `json:"mapState,omitempty"`
}

func (gameState *GameState) DrawCube(randProvider RandProvider) (Color, error) {
	var total int = 0
	for _, count := range gameState.CubeBag {
		total += count
	}
	if total == 0 {
		return NONE_COLOR, nil
	}
	val, err := randProvider.RandN(total)
	if err != nil {
		return NONE_COLOR, fmt.Errorf("failed to get random number: %v", err)
	}
	result := NONE_COLOR
	total = 0
	for color, count := range gameState.CubeBag {
		total += count
		if val < total {
			result = color
			break
		}
	}

	if prior, ok := gameState.CubeBag[result]; ok {
		if prior <= 0 {
			return NONE_COLOR, fmt.Errorf("internal error: picked color with no cubes")
		}
		gameState.CubeBag[result] = prior - 1
	} else {
		return NONE_COLOR, fmt.Errorf("internal error: picked invalid color: %v", result)
	}

	return result, nil
}
