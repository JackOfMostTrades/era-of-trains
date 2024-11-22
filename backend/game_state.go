package main

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Direction int

const (
	SOUTH Direction = iota
	SOUTH_WEST
	NORTH_WEST
	NORTH
	NORTH_EAST
	SOUTH_EAST
)

type Link struct {
	SourceHex       Coordinate  `json:"sourceHex"`
	SourceDirection Direction   `json:"sourceDirection"`
	Steps           []Direction `json:"steps"`
}

type Urbanization struct {
	Hex  Coordinate `json:"hex"`
	City string     `json:"city"`
}

type GameState struct {
	ActivePlayer  string            `json:"activePlayer"`
	PlayerOrder   []string          `json:"playerOrder"`
	PlayerShares  map[string]int    `json:"playerShares"`
	PlayerLoco    map[string]int    `json:"playerLoco"`
	PlayerIncome  map[string]int    `json:"playerIncome"`
	PlayerActions map[string]string `json:"playerActions"`
	PlayerCash    map[string]int    `json:"playerCash"`

	// Map from player ID to their last bid
	AuctionState map[string]int `json:"auctionState"`

	// 1 => shares, 2=>auction, 3=>choosing special actions, 4=>building, 5=>moving goods, 6=>moving goods (round 2)
	// (income and expenses has no user agency, so game state will never be in that phase)
	// 7 => goods growth (sometimes paused for user production action)
	GamePhase int `json:"gamePhase"`

	PlayerLinks   map[string][]*Link `json:"playerLinks"`
	Urbanizations []*Urbanization    `json:"urbanizations"`
}
