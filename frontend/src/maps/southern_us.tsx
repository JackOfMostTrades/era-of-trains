import {BasicMap} from "./basic_map.tsx";
import {ReactNode} from "react";

class SouthernUS extends BasicMap {

    public getMapInfo(): ReactNode {
        return <>
            <p>The four "port cities" (light city 4, light city 5, dark city 4, dark city 5) all accept white cubes and white cubes must stop movement at those cities.
            Delivering white cubes to one of those cities provides a bonus of one additional income.</p>
            <p>On turns 1-4 an additional cube from the bag is added to Atlanta (dark city 3) during Goods Growth.</p>
            <p>Starting on turn 4, income reduction is doubled.</p>
            </>;
    }

    public static fromJson(src: any): SouthernUS {
        let map = new SouthernUS();
        map.initializeFromJson(src);
        return map;
    }
}

export default SouthernUS
