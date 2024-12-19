package main

import (
	"encoding/json"
	"os"
	"strconv"
)

type Color int

const (
	NONE_COLOR Color = iota
	BLACK
	RED
	YELLOW
	BLUE
	PURPLE
)

func (c Color) String() string {
	switch c {
	case NONE_COLOR:
		return "NONE"
	case BLACK:
		return "BLACK"
	case RED:
		return "RED"
	case YELLOW:
		return "YELLOW"
	case BLUE:
		return "BLUE"
	case PURPLE:
		return "PURPLE"
	}
	return strconv.Itoa(int(c))
}

type BasicCity struct {
	Color      Color      `json:"color"`
	Coordinate Coordinate `json:"coordinate"`
	// Numbers for goods growth, 0-5 for white, 6-11 for black
	GoodsGrowth []int `json:"goodsGrowth"`
}

type StartingCubeSpec struct {
	Number     int        `json:"number"`
	Coordinate Coordinate `json:"coordinate"`
}

type BasicMap struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	// Rectangular array height*width in size (y dimension is first)
	Hexes         [][]HexType        `json:"hexes"`
	Cities        []BasicCity        `json:"cities"`
	StartingCubes []StartingCubeSpec `json:"startingCubes"`
}

type HexType int

const (
	OFFBOARD_HEX_TYPE HexType = iota
	WATER_HEX_TYPE
	PLAINS_HEX_TYPE
	RIVER_HEX_TYPE
	MOUNTAIN_HEX_TYPE
	TOWN_HEX_TYPE
	CITY_HEX_TYPE
)

func loadMaps() (map[string]*BasicMap, error) {
	maps := make(map[string]*BasicMap)
	rustBelt, err := loadBasicMap("maps/rust_belt.json")
	if err != nil {
		return nil, err
	}
	maps["rust_belt"] = rustBelt
	return maps, nil
}

func loadBasicMap(filename string) (*BasicMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := new(BasicMap)
	err = json.NewDecoder(f).Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
