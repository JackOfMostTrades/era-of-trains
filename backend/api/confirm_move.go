package api

import (
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/tiles"
)

type ActionName string

const (
	SharesActionName       ActionName = "shares"
	BidActionName          ActionName = "bid"
	ChooseActionName       ActionName = "choose_action"
	BuildActionName        ActionName = "build"
	MoveGoodsActionName    ActionName = "move_goods"
	ProduceGoodsActionName ActionName = "produce_goods"
)

type SharesAction struct {
	Amount int `json:"amount"`
}

type BidAction struct {
	Amount int `json:"amount"`
}

type ChooseAction struct {
	Action common.SpecialAction `json:"action"`
}

type TownPlacement struct {
	Track []common.Direction `json:"track"`
}

type TrackPlacement struct {
	Tile     tiles.TrackTile `json:"tile"`
	Rotation int             `json:"rotation"`
}

type TeleportLinkPlacement struct {
	Track common.Direction `json:"track"`
}

type BuildStep struct {
	Hex common.Coordinate `json:"hex"`
	// One of...
	// A=0, B=1, ...
	Urbanization          *int                   `json:"urbanization,omitempty"`
	TownPlacement         *TownPlacement         `json:"townPlacement,omitempty"`
	TrackPlacement        *TrackPlacement        `json:"trackPlacement,omitempty"`
	TeleportLinkPlacement *TeleportLinkPlacement `json:"teleportLinkPlacement,omitempty"`
}

type BuildAction struct {
	Steps []*BuildStep `json:"steps"`
}

type MoveGoodsAction struct {
	StartingLocation common.Coordinate  `json:"startingLocation"`
	Color            common.Color       `json:"color"`
	Path             []common.Direction `json:"path"`
	Loco             bool               `json:"loco"`
}

type ProduceGoodsAction struct {
	// List (corresponding the cubes in the same order as ProductionCubes in the game state) with X,Y coordinates
	// corresponding to which city (X) and which spot (Y) within that city
	Destinations []common.Coordinate `json:"destinations"`
}

type ConfirmMoveRequest struct {
	GameId             string              `json:"gameId"`
	ActionName         ActionName          `json:"actionName"`
	SharesAction       *SharesAction       `json:"sharesAction"`
	BidAction          *BidAction          `json:"bidAction"`
	ChooseAction       *ChooseAction       `json:"chooseAction"`
	BuildAction        *BuildAction        `json:"buildAction"`
	MoveGoodsAction    *MoveGoodsAction    `json:"moveGoodsAction"`
	ProduceGoodsAction *ProduceGoodsAction `json:"produceGoodsAction"`
}
type ConfirmMoveResponse struct {
}
