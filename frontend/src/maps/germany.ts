import {BasicMap} from "./basic_map.ts";
import {Color, Coordinate, GameState} from "../api/api.ts";
import {CityProperties} from "./index.ts";

class Germany extends BasicMap {

    private getPortNumber(hex: Coordinate): number {
        if (hex.x === 3 && hex.y === 1) {
            return 1;
        }
        if (hex.x === 0 && hex.y === 10) {
            return 2;
        }
        if (hex.x === 0 && hex.y === 16) {
            return 3;
        }
        if (hex.x === 0 && hex.y === 22) {
            return 4;
        }
        if (hex.x === 0 && hex.y === 28) {
            return 5;
        }
        if (hex.x === 6 && hex.y === 10) {
            return 6;
        }

        return 0;
    }

    public getCityProperties(gameState: GameState|undefined, hex: Coordinate): CityProperties | undefined {
        let portNumber = this.getPortNumber(hex);
        if (portNumber === 0) {
            return super.getCityProperties(gameState, hex);
        } else {
            if (gameState && gameState.mapState) {
                let portColors = gameState.mapState["portColors"] as Color[] | undefined;
                if (portColors) {
                    let color = portColors[portNumber - 1];
                    return {
                        color: color,
                        darkCity: false,
                        label: ""
                    };
                }
            }
        }

        return undefined;
    }

    public static fromJson(src: any): Germany {
        let map = new Germany();
        map.hexes = src.hexes;
        map.cities = src.cities;
        map.interurbanLinks = src.interurbanLinks;
        map.specialTrackPricing = src.specialTrackPricing;
        return map;
    }
}

export default Germany
