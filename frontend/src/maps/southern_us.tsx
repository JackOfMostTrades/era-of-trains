import {BasicMap, RIVER_COLOR} from "./basic_map.tsx";
import {ReactNode} from "react";

class SouthernUS extends BasicMap {

    public getMapInfo(): ReactNode {
        return <>
            <p>The four "port cities" (light city 4, light city 5, dark city 4, dark city 5) all accept white cubes and white cubes must stop movement at those cities.
            Delivering white cubes to one of those cities provides a bonus of one additional income.</p>
            <p>On turns 1-4 an additional cube from the bag is added to Atlanta (dark city 3) during Goods Growth.</p>
            <p>Income reduction is doubled on turn 4 (and only turn 4).</p>
            </>;
    }

    public getRiverLayer(): React.ReactNode {
        return <>
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 27.4255,7.5 c 7.772641,4.197346 2.134926,6.627625 5.561733,12.5 1.937257,3.319796 -4.445834,8.153133 -1.77277,10.89282 2.733873,2.802015 10.855839,4.958182 13.53154,6.60718"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 114.03,57.5 c 1.74858,0.919524 2.54508,3.074399 4.48472,3.696627 1.68516,0.540592 3.477,0.1791 5.21364,0.29782 1.1026,0.07538 2.40871,0.534835 2.71456,1.713534 0.26361,1.015903 -0.004,2.210702 0.75633,3.061973 0.81897,0.916922 2.74833,0.533874 4.15225,1.230046"
           />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 75.058,80 c -5.02e-4,2.120262 -0.907308,4.250715 -0.865816,6.605666 0.03389,1.923719 -0.118501,3.870139 0.722759,5.688634 0.88722,1.917844 1.171166,4.013136 1.147176,6.17992 -0.02063,1.86343 0.29313,3.9086 -0.790768,5.41864 C 73.614502,106.20112 75.057498,107.94218 75.058,110"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="none"
                d="m 40.416,80 c -5.02e-4,1.307933 -1.516315,4.037343 -1.686042,5.51697 -0.249849,2.178113 3.022692,4.921916 3.279805,7.263258 C 42.148719,95.580335 40.415498,98.442232 40.416,100"
            />
            <path
                stroke={RIVER_COLOR}
                strokeWidth="0.5"
                fill="#579ba8"
                d="m 13.062893,108.01891 c -0.534135,0.10755 -1.559442,0.12228 -1.707135,0.84615 -0.115226,0.56475 -0.09283,1.14937 -0.113381,1.72434 -0.02399,0.6713 0.438354,1.27801 1.045507,1.57469 0.745022,0.36405 1.7316,0.32203 2.292054,-0.35064 0.414051,-0.49695 0.966142,-1.05845 1.685003,-0.96804 0.591145,0.0743 1.356257,0.29067 1.804207,-0.22994 0.40624,-0.47214 0.02811,-1.13799 -0.270745,-1.57287 -0.633383,-0.92166 -1.482993,-1.71769 -2.250579,-1.65763 -0.728162,0.057 -1.668047,0.46946 -2.484931,0.63394 z"
            />
        </>
    }

    public static fromJson(src: any): SouthernUS {
        let map = new SouthernUS();
        map.initializeFromJson(src);
        return map;
    }
}

export default SouthernUS
