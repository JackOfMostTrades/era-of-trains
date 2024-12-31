import * as rustBeltRaw from "../../../backend/maps/rust_belt.json"
import * as southernUsRaw from "../../../backend/maps/southern_us.json"
import * as germanyRaw from "../../../backend/maps/germany.json"
import {Color, Coordinate, GameState} from "../api/api.ts";
import {BasicMap, InterurbanLink} from "./basic_map.tsx";
import Germany from "./germany.tsx";
import {ReactNode} from "react";
import SouthernUS from "./southern_us.tsx";
import RustBelt from "./rust_belt.tsx";

export enum HexType {
    OFFBOARD  = 0,
    WATER ,
    PLAINS ,
    RIVER ,
    MOUNTAIN ,
    TOWN ,
    CITY ,
    HILLS
}

export interface CityProperties {
    label: string
    color: Color
    darkCity: boolean
}

export interface GameMap {
    getWidth(): number;
    getHeight(): number;
    getHexType(hex: Coordinate): HexType;
    getCityProperties(gameState: GameState|undefined, hex: Coordinate): CityProperties|undefined;
    getCityColor(goodsGrowthNumber: number): Color;
    getInterurbanLinks(): InterurbanLink[];
    getSpecialTrackPricing(hex: Coordinate): number|undefined;
    getTurnLimit(playerCount: number): number
    getBuildLimit(gameState: GameState|undefined, player: string): number;
    getMapInfo(): ReactNode;
    getRiverLayer(): ReactNode;
}

const rustBelt = RustBelt.fromJson(rustBeltRaw);
const southernUs = SouthernUS.fromJson(southernUsRaw);
const germany = Germany.fromJson(germanyRaw);

export const maps: { [mapName: string]: BasicMap } = {
    "rust_belt": rustBelt,
    "southern_us": southernUs,
    "germany": germany,
}
