import {ReactNode} from "react";
import {BuildAction, Color, Coordinate, Direction, GameState} from "../api/api.ts";
import {CityProperties, GameMap, HexType} from "./index.tsx";

interface BasicMapHex {
    type: HexType;
    name?: string;
    cityColor?: Color;
    goodsGrowth?: number[];
    startingCubeCount?: number;
    cost?: number;
    mapData?: { [key: string]: any }
}

export interface TeleportLinkEdge {
    hex: Coordinate,
    direction: Direction
}

export interface TeleportLink {
    left: TeleportLinkEdge;
    right: TeleportLinkEdge;
    cost: number;
    costLocation: Coordinate;
    costLocationEdge: Direction|-1;
}

export const RIVER_COLOR = "#009bb2";

export class BasicMap implements GameMap {
    protected hexes: BasicMapHex[][] = [];
    protected teleportLinks: TeleportLink[] = [];

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
        return this.hexes[hex.y][hex.x].type;
    }

    public getLocationName(hex: Coordinate): string|undefined {
        return this.hexes[hex.y][hex.x].name;
    }

    public getCityProperties(_: GameState|undefined, hex: Coordinate): CityProperties|undefined {
        let city = this.hexes[hex.y][hex.x];
        if (city.type != HexType.CITY) {
            return undefined;
        }

        let label: string;
        if (city.goodsGrowth) {
            label = city.goodsGrowth.map(n => (n % 6) + 1).join(',');
        } else {
            label = "";
        }
        let color = city.cityColor || Color.NONE;
        let darkCity = false;
        if (city.goodsGrowth) {
            for (let goodsGrowth of city.goodsGrowth) {
                if (goodsGrowth >= 6) {
                    darkCity = true;
                    break;
                }
            }
        }
        return {
            label: label,
            color: color,
            darkCity: darkCity
        };

        return undefined;
    }

    public getCityColor(goodsGrowthNumber: number): Color {
        for (let y = 0; y < this.hexes.length; y++) {
            for (let x = 0; x < this.hexes[y].length; x++) {
                let city = this.hexes[y][x];
                if (city.goodsGrowth && city.goodsGrowth.indexOf(goodsGrowthNumber) !== -1) {
                    if (city.cityColor) {
                        return city.cityColor;
                    }
                    return Color.NONE;
                }
            }
        }
        return Color.NONE;
    }

    public getTeleportLinks(_gameState: GameState|undefined, _pendingBuildAction: BuildAction|undefined): TeleportLink[] {
        return this.teleportLinks || [];
    }

    public getSpecialTrackPricing(hex: Coordinate): number|undefined {
        let h = this.hexes[hex.y][hex.x];
        if (h.cost) {
            return h.cost;
        }
        return undefined;
    }

    public getTurnLimit(playerCount: number): number {
        if (playerCount === 6) {
            return 6
        } else if (playerCount === 5) {
            return 7
        } else if (playerCount === 4) {
            return 8
        }
        return 10
    }

    public getSharesLimit(): number {
        return 15;
    }

    public getBuildLimit(gameState: GameState | undefined, player: string): number {
        if (gameState && gameState.playerActions[player] === 'engineer') {
            return 4;
        }
        return 3;
    }

    public getMapInfo(): ReactNode {
        return null;
    }

    public getRiverLayer(): ReactNode {
        return null;
    }

    protected initializeFromJson(src: any) {
        this.hexes = src.hexes;
        this.teleportLinks = src.teleportLinks;
    }

    public static fromJson(src: any): BasicMap {
        let map = new BasicMap();
        map.initializeFromJson(src);
        return map;
    }
}
