import {BasicMap, RIVER_COLOR} from "./basic_map.tsx";
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
            } else {
                return {
                    color: Color.NONE,
                    darkCity: false,
                    label: "",
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

    public getRiverLayer(): React.ReactNode {
        return <>
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 36.0855,32.5 c 1.501231,0.855182 2.616603,2.447521 4.274327,2.988715 1.410079,0.460346 2.942767,0.349618 4.405895,0.628557 1.356812,0.258671 2.910472,0.453482 3.780758,1.679393 0.729351,1.027386 1.042395,2.000991 1.119959,2.965238 C 49.778084,42.149831 49.076498,43.49849 49.077,45"
                />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 49.077,55 c -5.02e-4,0.473964 -0.825598,1.46352 -0.219184,2.106739 1.641783,1.741431 1.667173,3.195181 0.665397,4.754605 C 48.767743,63.037353 49.076498,63.709329 49.077,65"
                />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 92.379,30 c -5.02e-4,2.180078 -0.162055,3.536692 0.906198,4.759142 1.101021,1.259949 2.308865,2.087908 3.423796,2.740856"
                />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 10.1045,67.5 c 1.012679,0.590773 1.768381,1.662526 3.038521,1.694182 1.788168,0.09563 3.362196,-0.902991 4.812724,-1.818867 1.551226,-0.979625 3.242424,-1.964585 5.122677,-1.962069 2.516078,0.0034 3.588106,1.630968 4.346578,2.086751"
                />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 23.095,80 c -5.02e-4,1.42244 -0.739539,3.646947 -0.416803,5 0.515834,2.162612 5.7128,1.123498 7.654029,3.712124 1.665889,2.143613 2.345447,5.127882 1.51869,7.74047 -0.546186,1.553126 -1.865901,2.978359 -1.463168,4.734936 0.255608,1.33717 1.36775,2.48299 1.368252,3.81247"
                />
        </>
    }

    public static fromJson(src: any): Germany {
        let map = new Germany();
        map.initializeFromJson(src);
        return map;
    }
}

export default Germany
