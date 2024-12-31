import {BasicMap, RIVER_COLOR, TeleportLink} from "./basic_map.tsx";
import {ReactNode} from "react";
import {BuildAction, GameState, Urbanization} from "../api/api.ts";

class Scotland extends BasicMap {

    public getMapInfo(): ReactNode {
        return <>
            <p>In a two-player game, if a player takes Turn Order Pass then the next auction is skipped and that player
                automatically becomes first player.</p>
            <p>Both towns on either side of the ferry links need to be urbanized before the ferry link can be built.</p>
            <p>Four dice are rolled for each of the light and dark goods growth phases.</p>
            <p>The game lasts 8 turns.</p>
        </>;
    }

    public getTeleportLinks(gameState: GameState|undefined, pendingBuildAction: BuildAction|undefined): TeleportLink[] {
        let teleportLinks = super.getTeleportLinks(gameState, pendingBuildAction).slice();
        let urbs: Urbanization[] = [];
        if (pendingBuildAction && pendingBuildAction.urbanization) {
            urbs.push(pendingBuildAction.urbanization);
        }
        if (gameState && gameState.urbanizations) {
            for (let urb of gameState.urbanizations) {
                urbs.push(urb);
            }
        }
        for (let urb of urbs) {
            if (urb.hex.x === 0 && urb.hex.y === 13) {
                teleportLinks.push({
                    left: {
                        hex: {x: 0, y: 13},
                        direction: 1,
                    },
                    right: {
                        hex: {x: 1, y: 12},
                        direction: 4,
                    },
                    cost: 2,
                    costLocation: {x: 0, y: 13},
                    costLocationEdge: 1,
                });
            }
        }
        return teleportLinks;
    }

    public getTurnLimit(_: number): number {
        return 8;
    }

    public getRiverLayer(): React.ReactNode {
        return <>
            <path
                stroke='#e1e1e1'
                strokeWidth="2"
                fill="none"
                d="m 5.774,80 c 0,-1.054479 0.036391,-3.045201 0.1787952,-3.80695 0.2684631,-1.436074 0.1173793,-0.983897 1.133226,-1.863262 C 7.6807276,73.814982 9.1679842,72.995733 10.1045,72.5"
            />
            <path
                stroke='#e1e1e1'
                strokeWidth="2"
                fill="none"
                d="m 10.104,12.5 c 1.053038,-0.620078 2.503555,-1.775293 3.182308,-2.18937 0.633793,-0.3866482 1.496811,-0.3251727 2.168164,0.04622 0.660037,0.36513 2.160061,1.46592 3.310028,2.143152"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="M 29.372782,9.5289626 C 31.755498,7.5 34.616246,7.6881987 36.138404,10.853153 c 0.870385,1.913043 1.408109,3.782363 3.142825,5.120017 0.927553,0.617685 2.138267,1.102768 3.21365,0.534504 C 43.284142,16.090605 43.987528,16.992533 44.746,17.5"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 18.7645,37.5 c 3.888203,-4.491103 5.364709,1.146232 7.699371,-3.11811 0.955458,-1.745177 0.09541,-3.766008 3.383259,-4.864222 2.913171,-0.973064 3.248823,2.09532 4.86177,1.49991 1.809123,-0.667827 1.919053,-3.286526 3.536202,-4.709548"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 42.585622,23.770032 c 1.120791,-0.835447 1.076177,-0.505816 2.155486,-1.25859"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 70.728,37.5 c -2.410099,1.206159 -0.160127,2.03655 -2.03643,3.565703 -0.990234,0.807024 -2.490461,0.821172 -3.410388,0.734283 C 64.285271,41.70592 63.502205,40.953938 62.8715,40.22795"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 53.407,52.5 c -1.81876,-1.226831 -2.071371,-2.17042 -3.113626,-3.132881 -1.185838,-1.095052 -2.888324,-0.746848 -4.35888,-0.43371 -2.361323,0.430117 -4.661578,-0.515004 -6.735542,-1.567619 -1.150955,-0.671812 -2.343621,-1.854622 -1.471164,-3.289869 0.70838,-1.16533 1.959769,-1.57257 3.531632,0.572665"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 44.746,62.5 c -1.830499,1.225 -0.251058,3.776774 -2.788436,4.56518 -2.24438,0.697367 -3.58979,-0.840732 -3.43178,-3.064144 0.181862,-2.559038 -3.276583,-2.222688 -5.25114,-1.569215 -1.520591,0.503234 -5.123022,-0.406049 -4.219299,-2.708962 0.847182,-2.158832 -1.414803,-3.250258 -3.08288,-2.378117 -1.562163,0.816765 -3.758284,0.287594 -3.516586,-1.876739 0.139781,-1.251693 -1.391756,-2.921965 -2.710315,-2.321957"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 9.7745544,61.889139 c 1.2686986,-0.917969 2.7182806,-0.331854 4.0082166,0.288821 1.209993,0.582209 2.004865,-1.820091 3.418281,-1.124912 C 17.778548,61.46417 18.205393,62.117921 18.765,62.5"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 22.938276,75.200215 c 1.782013,-0.249586 3.193395,0.843929 4.252637,2.253087 1.236822,1.645401 3.515875,1.397685 5.403192,1.918132 C 34.8735,80 34.993271,81.75 36.0855,82.5"
            />
        </>
    }

    public static fromJson(src: any): Scotland {
        let map = new Scotland();
        map.initializeFromJson(src);
        return map;
    }
}

export default Scotland
