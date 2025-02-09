package maps

import (
	"encoding/json"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/tiles"
	"os"
	"slices"

	"github.com/JackOfMostTrades/eot/backend/common"
)

type teleportLinkEdge struct {
	Hex       common.Coordinate `json:"hex"`
	Direction common.Direction  `json:"direction"`
}

type teleportLink struct {
	Left             *teleportLinkEdge `json:"left"`
	Right            *teleportLinkEdge `json:"right"`
	Cost             int               `json:"cost"`
	CostLocation     common.Coordinate `json:"costLocation"`
	CostLocationEdge common.Direction  `json:"costLocationEdge"`
}

type basicMapHex struct {
	HexType           HexType                `json:"type"`
	Name              string                 `json:"name,omitempty"`
	CityColor         common.Color           `json:"cityColor,omitempty"`
	GoodsGrowth       []int                  `json:"goodsGrowth,omitempty"`
	StartingCubeCount int                    `json:"startingCubeCount,omitempty"`
	Cost              int                    `json:"cost,omitempty"`
	MapData           map[string]interface{} `json:"mapData,omitempty"`
}

type basicMap struct {
	*AbstractGameMapImpl

	// Rectangular array height*width in size (y dimension is first)
	Hexes         [][]*basicMapHex `json:"hexes"`
	TeleportLinks []teleportLink   `json:"teleportLinks"`
}

func (b *basicMap) PopulateStartingCubes(gameState *common.GameState, randProvider common.RandProvider) error {
	for y := 0; y < len(b.Hexes); y++ {
		for x := 0; x < len(b.Hexes[y]); x++ {
			for i := 0; i < b.Hexes[y][x].StartingCubeCount; i++ {
				cube, err := gameState.DrawCube(randProvider)
				if err != nil {
					return fmt.Errorf("failed to draw cube: %v", err)
				}
				gameState.Cubes = append(gameState.Cubes, &common.BoardCube{
					Color: cube,
					Hex:   common.Coordinate{X: x, Y: y},
				})
			}
		}
	}
	return nil
}

func (b *basicMap) GetCityColorForHex(gameState *common.GameState, hex common.Coordinate) common.Color {
	return b.Hexes[hex.Y][hex.X].CityColor
}

func (b *basicMap) GetCityHexForGoodsGrowth(goodsGrowth int) common.Coordinate {
	for y := 0; y < len(b.Hexes); y++ {
		for x := 0; x < len(b.Hexes[y]); x++ {
			hex := b.Hexes[y][x]
			if slices.Index(hex.GoodsGrowth, goodsGrowth) != -1 {
				return common.Coordinate{X: x, Y: y}
			}
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
	return b.Hexes[hex.Y][hex.X].HexType
}

func (b *basicMap) GetTrackBuildCost(gameState *common.GameState, player string, hexType HexType, hex common.Coordinate, trackType tiles.TrackType, isUpgrade bool) (int, error) {
	if !isUpgrade {
		boardHex := b.Hexes[hex.Y][hex.X]
		if boardHex.Cost != 0 {
			return boardHex.Cost, nil
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
