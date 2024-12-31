package maps

import (
	"encoding/json"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"os"
	"slices"
)

type basicCity struct {
	Color      common.Color      `json:"color"`
	Coordinate common.Coordinate `json:"coordinate"`
	// Numbers for goods growth, 0-5 for white, 6-11 for black
	GoodsGrowth []int `json:"goodsGrowth"`
}

type startingCubeSpec struct {
	Number     int               `json:"number"`
	Coordinate common.Coordinate `json:"coordinate"`
}

type specialTrackPricing struct {
	Cost int               `json:"cost"`
	Hex  common.Coordinate `json:"hex"`
}

type teleportLink struct {
	Left             *teleportLinkEdge `json:"left"`
	Right            *teleportLinkEdge `json:"right"`
	Cost             int               `json:"cost"`
	CostLocation     common.Coordinate `json:"costLocation"`
	CostLocationEdge common.Direction  `json:"costLocationEdge"`
}

type teleportLinkEdge struct {
	Hex       common.Coordinate `json:"hex"`
	Direction common.Direction  `json:"direction"`
}

type basicMap struct {
	*AbstractGameMapImpl

	// Rectangular array height*width in size (y dimension is first)
	Hexes         [][]HexType        `json:"hexes"`
	Cities        []basicCity        `json:"cities"`
	StartingCubes []startingCubeSpec `json:"startingCubes"`
	TeleportLinks []teleportLink     `json:"teleportLinks"`
	// Hexes with unusual track costs
	SpecialTrackPricing []specialTrackPricing `json:"specialTrackPricing,omitempty"`
}

func (b *basicMap) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	for _, startingCubeSpec := range b.StartingCubes {
		for i := 0; i < startingCubeSpec.Number; i++ {
			cube, err := gameState.DrawCube(randProvider)
			if err != nil {
				return fmt.Errorf("failed to draw cube: %v", err)
			}
			gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
				Color: cube,
				Hex:   startingCubeSpec.Coordinate,
			})
		}
	}
	return nil
}

func (b *basicMap) GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color {
	for _, city := range b.Cities {
		if city.Coordinate.Equals(hex) {
			return city.Color
		}
	}
	return common.NONE_COLOR
}

func (b *basicMap) GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate {
	for _, city := range b.Cities {
		if slices.Index(city.GoodsGrowth, goodsGrowth) != -1 {
			return city.Coordinate
		}
	}
	return common.Coordinate{X: -1, Y: -1}
}

func (b *basicMap) GetWidth() int {
	return len(b.Hexes[0])
}

func (b *basicMap) GetHeight() int {
	return len(b.Hexes)
}

func (b *basicMap) GetHexType(hex common.Coordinate) HexType {
	if hex.X < 0 || hex.Y < 0 || hex.Y >= len(b.Hexes) || hex.X >= len(b.Hexes[hex.Y]) {
		return OFFBOARD_HEX_TYPE
	}
	return b.Hexes[hex.Y][hex.X]
}

func (b *basicMap) GetTrackBuildCost(gameState *common.GameState, player string, hexType HexType, hex common.Coordinate, trackType common.TrackType, isUpgrade bool) (int, error) {
	if !isUpgrade {
		for _, pricing := range b.SpecialTrackPricing {
			if pricing.Hex.X == hex.X && pricing.Hex.Y == hex.Y {
				cost := pricing.Cost
				switch trackType {
				case common.SIMPLE_TRACK_TYPE:
					break
				case common.COMPLEX_COEXISTING_TRACK_TYPE:
					cost += 1
				case common.COMPLEX_CROSSING_TRACK_TYPE:
					cost += 2
				default:
					panic(fmt.Errorf("Unhandled track type: %d", trackType))
				}
				return cost, nil
			}
		}
	}

	return b.AbstractGameMapImpl.GetTrackBuildCost(gameState, player, hexType, hex, trackType, isUpgrade)
}

func (b *basicMap) GetTeleportLinkBuildCost(gameState *common.GameState, player string, hex common.Coordinate, direction common.Direction) int {
	for _, link := range b.TeleportLinks {
		if (link.Left.Hex == hex && link.Left.Direction == direction) ||
			(link.Right.Hex == hex && link.Right.Direction == direction) {
			return link.Cost
		}
	}
	return 0
}

func (b *basicMap) GetTeleportLink(gameState *common.GameState, src common.Coordinate, direction common.Direction) (*common.Coordinate, common.Direction) {
	for _, teleportLink := range b.TeleportLinks {
		if teleportLink.Left.Hex == src && teleportLink.Left.Direction == direction {
			dest := teleportLink.Right.Hex
			return &dest, teleportLink.Right.Direction
		}
		if teleportLink.Right.Hex == src && teleportLink.Right.Direction == direction {
			dest := teleportLink.Left.Hex
			return &dest, teleportLink.Left.Direction
		}
	}
	return nil, 0
}

func loadBasicMap(filename string) (*basicMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := new(basicMap)
	err = json.NewDecoder(f).Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
