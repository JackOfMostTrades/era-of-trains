import {BasicMap} from "./basic_map.tsx";
import {Color, Coordinate, GameState} from "../api/api.ts";
import {CityProperties} from "./index.tsx";
import {ReactNode} from "react";

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

    public getBuildLimit(_gameState: GameState | undefined, _player: string): number {
        return 3;
    }

    public getMapInfo(): ReactNode {
        return <>
            <p>The color of the six unnumbered border cities is randomly determined during setup. Cubes can never pass through these cities (only end there).</p>
            <p>Engineer only allows 3 builds, but the cost of one track tile placement (the most expensive) is halved (rounded up).</p>
            <p>You cannot have any incomplete track at the end of your build action.</p>
            <p>Berlin (the unnumbered black city in the center) always receives one cube during Goods Growth.</p>
        </>;
    }

    public static fromJson(src: any): Germany {
        let map = new Germany();
        map.initializeFromJson(src);
        return map;
    }
}

export default Germany
