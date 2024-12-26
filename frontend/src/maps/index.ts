import * as rustBeltRaw from "../../../backend/maps/rust_belt.json"
import * as southernUsRaw from "../../../backend/maps/southern_us.json"
import * as germanyRaw from "../../../backend/maps/germany.json"
import {Color, Coordinate, GameState} from "../api/api.ts";
import {BasicMap} from "./basic_map.ts";
import Germany from "./germany.ts";

export enum HexType {
    OFFBOARD  = 0,
    WATER ,
    PLAINS ,
    RIVER ,
    MOUNTAIN ,
    TOWN ,
    CITY ,
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
    getSpecialTrackPricing(hex: Coordinate): number|undefined;
}

const rustBelt = BasicMap.fromJson(rustBeltRaw);
const southernUs = BasicMap.fromJson(southernUsRaw);
const germany = Germany.fromJson(germanyRaw);

export const maps: { [mapName: string]: BasicMap } = {
    "rust_belt": rustBelt,
    "southern_us": southernUs,
    "germany": germany,
}
