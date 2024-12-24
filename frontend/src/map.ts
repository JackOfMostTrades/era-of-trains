import * as rustBeltRaw from "../../backend/maps/rust_belt.json"
import * as southernUsRaw from "../../backend/maps/southern_us.json"
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

interface BasicCity {
    color: Color;
    coordinate: Coordinate;
    goodsGrowth: number[];
}

export interface CityProperties {
    label: string
    color: Color
    darkCity: boolean
}

export class BasicMap {
    private hexes: HexType[][] = [];
    private cities: BasicCity[] = [];

    public getWidth(): number {
        return this.hexes[0].length;
    }

    public getHeight(): number {
        return this.hexes.length;
    }

    public getHexType(hex: Coordinate): HexType {
        if (hex.x < 0 || hex.y < 0 || hex.y >= this.hexes.length || hex.x >= this.hexes[0].length) {
            return HexType.OFFBOARD;
        }
        return this.hexes[hex.y][hex.x];
    }

    public getCityProperties(hex: Coordinate): CityProperties|undefined {
        for (let city of this.cities) {
            if (city.coordinate.x === hex.x && city.coordinate.y === hex.y) {
                let label = city.goodsGrowth.map(n => (n%6)+1).join(',');
                let color = city.color;
                let darkCity = false;
                for (let goodsGrowth of city.goodsGrowth) {
                    if (goodsGrowth >= 6) {
                        darkCity = true;
                        break;
                    }
                }
                return {
                    label: label,
                    color: color,
                    darkCity: darkCity
                };
            }
        }

        return undefined;
    }

    public static fromJson(src: any): BasicMap {
        let map = new BasicMap();
        map.hexes = src.hexes;
        map.cities = src.cities;
        return map;
    }
}

const rustBelt = BasicMap.fromJson(rustBeltRaw);
const southernUs = BasicMap.fromJson(southernUsRaw);

const maps: { [mapName: string]: BasicMap } = {
    "rust_belt": rustBelt,
    "southern_us": southernUs,
}

export default maps;
