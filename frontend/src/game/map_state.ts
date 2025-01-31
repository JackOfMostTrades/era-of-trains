import {Coordinate, Direction, GameState} from "../api/api.ts";
import {GameMap, HexType} from "../maps";
import {applyDirection, oppositeDirection} from "../util.ts";

export interface Route {
    left: Direction;
    right: Direction;
    owner: string;
}

interface Dangler {
    from: Direction;
    owner: string;
}

interface TileState {
    routes: Array<Route>
    danglers: Array<Dangler>
    isTown: boolean;
    isCity: boolean;
}

export class MapTileState {
    private map: GameMap;
    private tileState: TileState[][];

    constructor(map: GameMap, gameState: GameState) {
        this.map = map;
        this.tileState = [];
        for (let y = 0; y < map.getHeight(); y++) {
            this.tileState.push([]);
            for (let x = 0; x < map.getWidth(); x++) {
                let hexType = map.getHexType({x: x, y: y});
                this.tileState[y].push({
                    routes: [],
                    danglers: [],
                    isTown: hexType === HexType.TOWN,
                    isCity: hexType === HexType.CITY,
                });
            }
        }

        if (gameState.urbanizations) {
            for (let urb of gameState.urbanizations) {
                this.tileState[urb.hex.y][urb.hex.x].isCity = true;
            }
        }
        if (gameState.links) {
            for (let link of gameState.links) {
                let hex = link.sourceHex;
                if (this.tileState[hex.y][hex.x].isTown) {
                    this.tileState[hex.y][hex.x].routes.push({
                        left: link.steps[0],
                        right: link.steps[0],
                        owner: link.owner,
                    });
                }

                for (let i = 1; i < link.steps.length; i++) {
                    hex = applyDirection(hex, link.steps[i-1]);
                    this.tileState[hex.y][hex.x].routes.push({
                        left: oppositeDirection(link.steps[i - 1]),
                        right: link.steps[i],
                        owner: link.owner,
                    });
                }

                if (!link.complete && link.steps.length > 1) {
                    this.tileState[hex.y][hex.x].danglers.push({
                        from: oppositeDirection(link.steps[link.steps.length-2]),
                        owner: link.owner
                    });
                }

                hex = applyDirection(hex, link.steps[link.steps.length-1]);
                if (this.tileState[hex.y][hex.x].isTown && link.complete) {
                    let dir = oppositeDirection(link.steps[link.steps.length-1]);
                    this.tileState[hex.y][hex.x].routes.push({
                        left: dir,
                        right: dir,
                        owner: link.owner,
                    });
                }
            }
        }
    }

    public isTown(coordinate: Coordinate): boolean {
        return this.tileState[coordinate.y][coordinate.x].isTown;
    }

    public isCity(coordinate: Coordinate): boolean {
        return this.tileState[coordinate.y][coordinate.x].isCity;
    }

    public getTileState(coordinate: Coordinate): TileState {
        return this.tileState[coordinate.y][coordinate.x];
    }

    public isValidTrackPlacement(coordinate: Coordinate,
                                 tile: TrackTile,
                                 rotation: Rotation,
                                 activePlayer: string): boolean {
        let tileState = this.tileState[coordinate.y][coordinate.x];
        if (tileState.isTown || tileState.isCity) {
            return false;
        }

        let newTrackTile = new MapTrackTile(tile, rotation);

        // Check that any routes that must be preserved are in fact preserved
        for (let oldRoute of tileState.routes) {
            let mustBePreserved = false;
            // Any routes owned by another player must be preserved
            if (oldRoute.owner && oldRoute.owner !== activePlayer) {
                mustBePreserved = true;
            }
            // Any routes that are not a dangler must be preserved
            let isDangler = false;
            for (let dangler of tileState.danglers) {
                if (dangler.from === oldRoute.left || dangler.from === oldRoute.right) {
                    isDangler = true;
                    break;
                }
            }
            if (!isDangler) {
                mustBePreserved = true;
            }

            // Check that it is preserved if it must be preserved.
            if (mustBePreserved) {
                let isPreserved = false;
                for (let newRoute of newTrackTile.getRoutes()) {
                    if ((newRoute[0] === oldRoute.left && newRoute[1] === oldRoute.right)
                            || (newRoute[0] === oldRoute.right && newRoute[1] === oldRoute.left)) {
                        isPreserved = true;
                        break;
                    }
                }
                if (!isPreserved) {
                    return false;
                }
            }
        }

        // Check that there is a track from each dangler direction
        for (let dangler of tileState.danglers) {
            let hasExtension = false;
            for (let newRoute of newTrackTile.getRoutes()) {
                if (newRoute[0] === dangler.from || newRoute[1] === dangler.from) {
                    hasExtension = true;
                }
            }
            if (!hasExtension) {
                return false;
            }
        }

        // Every exit must lead to a passable terrain
        for (let newRoute of newTrackTile.getRoutes()) {
            for (let dir of newRoute) {
                let newHex = applyDirection(coordinate, dir);
                let newHexType = this.map.getHexType(newHex);
                if (newHexType === HexType.WATER || newHexType === HexType.OFFBOARD) {
                    return false;
                }
            }
        }

        return true;
    }

    public getNewRoutes(coordinate: Coordinate,
                        tile: TrackTile,
                        rotation: Rotation): Array<[Direction, Direction]> {
        let tileState = this.tileState[coordinate.y][coordinate.x];
        let newTrackTile = new MapTrackTile(tile, rotation);

        let newRoutes: Array<[Direction, Direction]> = [];
        for (let route of newTrackTile.getRoutes()) {
            let isOldRoute = false;
            for (let oldRoute of tileState.routes) {
                if ((route[0] === oldRoute.left && route[1] == oldRoute.right)
                        || (route[0] === oldRoute.right && route[1] == oldRoute.left)) {
                    isOldRoute = true;
                    break;
                }
            }
            if (!isOldRoute) {
                let isRedirect = false;
                for (let dangler of tileState.danglers) {
                    if ((dangler.from === route[0] || dangler.from === route[1])) {
                        isRedirect = true;
                        break;
                    }
                }
                if (!isRedirect) {
                    newRoutes.push(route);
                }
            }
        }

        return newRoutes;
    }

    public getRedirectedRoutes(coordinate: Coordinate,
                               tile: TrackTile,
                               rotation: Rotation): Array<[Direction, Direction]> {
        let tileState = this.tileState[coordinate.y][coordinate.x];
        let newTrackTile = new MapTrackTile(tile, rotation);

        let redirectedRoutes: Array<[Direction, Direction]> = [];
        for (let route of newTrackTile.getRoutes()) {
            let isOldRoute = false;
            for (let oldRoute of tileState.routes) {
                if ((route[0] === oldRoute.left && route[1] == oldRoute.right)
                    || (route[0] === oldRoute.right && route[1] == oldRoute.left)) {
                    isOldRoute = true;
                    break;
                }
            }
            if (isOldRoute) {
                continue;
            }

            for (let dangler of tileState.danglers) {
                if (dangler.from === route[0]) {
                    redirectedRoutes.push([dangler.from, route[1]]);
                }
                if (dangler.from === route[1]) {
                    redirectedRoutes.push([dangler.from, route[0]]);
                }
            }
        }

        return redirectedRoutes;
    }

    public isValidTownPlacement(coordinate: Coordinate,
                                tile: BasicTownTile,
                                rotation: Rotation): boolean {
        let tileState = this.tileState[coordinate.y][coordinate.x];
        if (!tileState.isTown || tileState.isCity) {
            return false;
        }

        let newTownTile = new TownTile(tile, rotation);

        // All existing exists must be preserved.
        for (let oldExit of tileState.routes) {
            let isPreserved = false;
            for (let newExit of newTownTile.getExits()) {
                if (newExit === oldExit.left || newExit === oldExit.right) {
                    isPreserved = true;
                    break;
                }
            }
            if (!isPreserved) {
                return false;
            }
        }

        // Every exit must lead to a passable terrain
        for (let newExit of newTownTile.getExits()) {
            let newHex = applyDirection(coordinate, newExit);
            let newHexType = this.map.getHexType(newHex);
            if (newHexType === HexType.WATER || newHexType === HexType.OFFBOARD) {
                return false;
            }
        }

        return true;
    }

    public getNewTownExits(coordinate: Coordinate,
                           tile: BasicTownTile,
                           rotation: Rotation): Array<Direction> {
        let tileState = this.tileState[coordinate.y][coordinate.x];
        let newTownTile = new TownTile(tile, rotation);

        let newExits: Array<Direction> = [];
        for (let exit of newTownTile.getExits()) {
            let isOldExit = false;
            for (let oldRoute of tileState.routes) {
                if (oldRoute.left === exit || oldRoute.right === exit) {
                    isOldExit = true;
                    break;
                }
            }
            if (!isOldExit) {
                newExits.push(exit);
            }
        }

        return newExits;
    }
}

// This needs to match the trackTile enum in component_limit_checks.go
export enum TrackTile {
    // Simple
    STRAIGHT_TRACK_TILE = 1,
    SHARP_CURVE_TRACK_TILE,
    GENTLE_CURVE_TRACK_TILE,

    // Complex crossing
    BOW_AND_ARROW_TRACK_TILE,
    TWO_GENTLE_TRACK_TILE,
    TWO_STRAIGHT_TRACK_TILE,

    // Complex coexist
    BASEBALL_TRACK_TILE,
    LEFT_GENTLE_AND_SHARP_TRACK_TILE,
    RIGHT_GENTLE_AND_SHARP_TRACK_TILE,
    STRAIGHT_AND_SHARP_TRACK_TILE,
}
export const TrackTiles: TrackTile[] =
    [TrackTile.STRAIGHT_TRACK_TILE, TrackTile.GENTLE_CURVE_TRACK_TILE, TrackTile.SHARP_CURVE_TRACK_TILE,
        TrackTile.BOW_AND_ARROW_TRACK_TILE, TrackTile.TWO_GENTLE_TRACK_TILE, TrackTile.TWO_STRAIGHT_TRACK_TILE,
        TrackTile.BASEBALL_TRACK_TILE, TrackTile.LEFT_GENTLE_AND_SHARP_TRACK_TILE, TrackTile.RIGHT_GENTLE_AND_SHARP_TRACK_TILE, TrackTile.STRAIGHT_AND_SHARP_TRACK_TILE];

const TRACK_TILE_ROUTES: Map<TrackTile, Array<[Direction, Direction]>> = new Map([
    [TrackTile.STRAIGHT_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH]]],
    [TrackTile.SHARP_CURVE_TRACK_TILE, [[Direction.SOUTH_EAST, Direction.SOUTH]]],
    [TrackTile.GENTLE_CURVE_TRACK_TILE, [[Direction.NORTH_EAST, Direction.SOUTH]]],

    [TrackTile.BOW_AND_ARROW_TRACK_TILE, [[Direction.NORTH_EAST, Direction.SOUTH], [Direction.SOUTH_EAST, Direction.NORTH_WEST]]],
    [TrackTile.TWO_GENTLE_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH_EAST], [Direction.NORTH_EAST, Direction.SOUTH]]],
    [TrackTile.TWO_STRAIGHT_TRACK_TILE, [[Direction.NORTH_EAST, Direction.SOUTH_WEST], [Direction.SOUTH_EAST, Direction.NORTH_WEST]]],

    [TrackTile.BASEBALL_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH_WEST], [Direction.NORTH_EAST, Direction.SOUTH]]],
    [TrackTile.LEFT_GENTLE_AND_SHARP_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH_EAST], [Direction.SOUTH_WEST, Direction.NORTH_WEST]]],
    [TrackTile.RIGHT_GENTLE_AND_SHARP_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH_WEST], [Direction.NORTH_EAST, Direction.SOUTH_EAST]]],
    [TrackTile.STRAIGHT_AND_SHARP_TRACK_TILE, [[Direction.NORTH, Direction.SOUTH], [Direction.SOUTH_WEST, Direction.NORTH_WEST]]],
]);

export type Rotation = 0 | 1 | 2 | 3 | 4 | 5;

export class MapTrackTile {
    private tile: TrackTile;
    private rotation: Rotation;

    constructor(tile: TrackTile, rotation: Rotation) {
        this.tile = tile;
        this.rotation = rotation;
    }

    public getRoutes(): Array<[Direction, Direction]> {
        let routes: Array<[Direction, Direction]>|undefined = TRACK_TILE_ROUTES.get(this.tile);
        if (!routes) {
            throw Error("Invalid tile: " + this.tile);
        }
        let rotatedRoutes: Array<[Direction, Direction]> = [];
        for (let route of routes) {
            let rotatedRoute: [Direction, Direction] = [
                (route[0]+this.rotation)%6 as Direction, (route[1]+this.rotation)%6 as Direction
            ];
            rotatedRoutes.push(rotatedRoute);
        }
        return rotatedRoutes;
    }
}

export type BasicTownTile = Array<Direction>;
export const TOWN_TILES: BasicTownTile[] = [
    [Direction.NORTH],

    [Direction.NORTH, Direction.NORTH_EAST],
    [Direction.NORTH, Direction.SOUTH_EAST],
    [Direction.NORTH, Direction.SOUTH],

    [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH_EAST],
    [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH],
    [Direction.NORTH, Direction.SOUTH_EAST, Direction.SOUTH],
    [Direction.NORTH, Direction.SOUTH_WEST, Direction.SOUTH_EAST],

    [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH_EAST, Direction.SOUTH],
    [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH, Direction.SOUTH_WEST],
    [Direction.NORTH, Direction.SOUTH_WEST, Direction.SOUTH, Direction.SOUTH_EAST],
]

export class TownTile {
    private tile: BasicTownTile;
    private rotation: Rotation;

    constructor(tile: BasicTownTile, rotation: Rotation) {
        this.tile = tile;
        this.rotation = rotation;
    }

    public getExits(): Array<Direction> {
        let routes: Direction[] = [];
        for (let route of this.tile) {
            let rotatedRoute: Direction = (route+this.rotation)%6
            routes.push(rotatedRoute);
        }
        return routes;
    }

    public static fromId(id: number): TownTile {
        let trackTypeId = Math.floor(id/6);
        let rotation: Rotation = (id%6) as Rotation;
        return new TownTile(TOWN_TILES[trackTypeId], rotation);
    }
}
