import {ReactNode} from "react";
import {BasicMap} from "./basic_map.tsx";
import {GameState} from "../api/api.ts";

class Australia extends BasicMap {

  public getMapInfo(): ReactNode {
    return <>
      <p>Plays 4-6, best 4-6.</p>
      <p>Engineering allows you to build one tile for free. (It does not grant you the usual extra build.)</p>
      <p>Urbanization limits you to only 2 builds.</p>
      <p>Delivering to Perth (the Blue white #1 city) earns a bonus 3 income per delivery.</p>
    </>;
  }

  public getBuildLimit(gameState: GameState | undefined, player: string): number {
    if (gameState && gameState.playerActions[player] === 'urbanization') {
      return 2;
    }
    return 3;
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
