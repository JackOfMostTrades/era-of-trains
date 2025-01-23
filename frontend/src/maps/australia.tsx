import {ReactNode} from "react";
import {BasicMap} from "./basic_map.tsx";

class Australia extends BasicMap {

  public getMapInfo(): ReactNode {
    return <>
      <p>Plays 4-6, best 4-6.</p>
      <p>Engineering allows you to build one tile for free.</p>
      <p>Urbanization limits you to only 2 builds.</p>
      <p>Delivering to Perth (the Blue white #1 city) earns a bonus 3 income per delivery.</p>
    </>;
  }

  public getRiverLayer(): React.ReactNode {
    return null;
  }

  public static fromJson(src: any): Australia {
    let map = new Australia();
    map.initializeFromJson(src);
    return map;
  }
}

export default Australia
