import * as rustBeltRaw from "../../backend/maps/rust_belt.json"
import {Color, Coordinate} from "./api/api.ts";

export enum HexType {
    OFFBOARD  = 0,
    WATER ,
    PLAINS ,
    RIVER ,
    MOUNTAIN ,
    TOWN ,
    CITY ,
}

export interface BasicCity {
    color: Color;
    coordinate: Coordinate;
    goodsGrowth: number[];
}

export class BasicMap {
    public width: number = 0;
    public height: number = 0;
    public hexes: HexType[][] = [];
    public cities: BasicCity[] = [];

    public static fromJson(src: any): BasicMap {
        let map = new BasicMap();
        map.width = src.width;
        map.height = src.height;
        map.hexes = src.hexes;
        map.cities = src.cities;
        return map;
    }
}

const rustBelt = BasicMap.fromJson(rustBeltRaw);

const maps: { [mapName: string]: BasicMap } = {
    "rust_belt": rustBelt,
}

export default maps;
