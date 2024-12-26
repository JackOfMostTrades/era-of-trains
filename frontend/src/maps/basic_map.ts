import {Color, Coordinate, Direction, GameState} from "../api/api.ts";
import {CityProperties, GameMap, HexType} from "./index.ts";

interface BasicCity {
    color: Color;
    coordinate: Coordinate;
    goodsGrowth: number[];
}

export interface InterurbanLink {
    cost: number;
    hex: Coordinate;
    direction: Direction;
}

interface SpecialTrackPricing {
    cost: number;
    hex: Coordinate;
}

export class BasicMap implements GameMap {
    protected hexes: HexType[][] = [];
    protected cities: BasicCity[] = [];
    protected interurbanLinks: InterurbanLink[] = [];
    protected specialTrackPricing: SpecialTrackPricing[] = [];

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

    public getCityProperties(_: GameState|undefined, hex: Coordinate): CityProperties|undefined {
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

    public getInterurbanLinks(): InterurbanLink[] {
        return this.interurbanLinks || [];
    }

    public getSpecialTrackPricing(hex: Coordinate): number|undefined {
        if (this.specialTrackPricing) {
            for (let pricing of this.specialTrackPricing) {
                if (pricing.hex.x === hex.x && pricing.hex.y === hex.y) {
                    return pricing.cost;
                }
            }
        }
        return undefined;
    }

    public static fromJson(src: any): BasicMap {
        let map = new BasicMap();
        map.hexes = src.hexes;
        map.cities = src.cities;
        map.interurbanLinks = src.interurbanLinks;
        map.specialTrackPricing = src.specialTrackPricing;
        return map;
    }
}
