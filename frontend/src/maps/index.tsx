import {ReactNode} from "react";
import * as australiaRaw from "../../../backend/maps/australia.json";
import * as germanyRaw from "../../../backend/maps/germany.json";
import * as rustBeltRaw from "../../../backend/maps/rust_belt.json";
import * as scotlandRaw from "../../../backend/maps/scotland.json";
import * as southernUsRaw from "../../../backend/maps/southern_us.json";
import {BuildAction, Color, Coordinate, GameState} from "../api/api.ts";
import Australia from "./australia.tsx";
import {BasicMap, TeleportLink} from "./basic_map.tsx";
import Germany from "./germany.tsx";
import RustBelt from "./rust_belt.tsx";
import Scotland from "./scotland.tsx";
import SouthernUS from "./southern_us.tsx";

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
    getLocationName(hex: Coordinate): string|undefined;
    getCityColor(goodsGrowthNumber: number): Color;
    getTeleportLinks(gameState: GameState|undefined, pendingBuildAction: BuildAction|undefined): TeleportLink[];
    getSpecialTrackPricing(hex: Coordinate): number|undefined;
    getTurnLimit(gameState: GameState|undefined, playerCount: number): number
    getSharesLimit(): number
    getBuildLimit(gameState: GameState|undefined, player: string): number;
    getMapInfo(): ReactNode;
    getRiverLayer(): ReactNode;
}

const rustBelt = RustBelt.fromJson(rustBeltRaw);
const southernUs = SouthernUS.fromJson(southernUsRaw);
const germany = Germany.fromJson(germanyRaw);
const scotland = Scotland.fromJson(scotlandRaw);
const australia = Australia.fromJson(australiaRaw);

export const maps: { [mapName: string]: BasicMap } = {
    "rust_belt": rustBelt,
    "southern_us": southernUs,
    "germany": germany,
    "scotland": scotland,
    "australia": australia,
}
