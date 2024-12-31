import {Coordinate, Direction, SpecialAction} from "./api/api.ts";
import {GameMap} from "./maps";

export function applyTeleport(map: GameMap, coordinate: Coordinate, direction: Direction): Coordinate|undefined {
    for (let teleportLink of map.getTeleportLinks()) {
        if (teleportLink.left.hex.x === coordinate.x && teleportLink.left.hex.y === coordinate.y && teleportLink.left.direction === direction) {
            return teleportLink.right.hex;
        }
        if (teleportLink.right.hex.x === coordinate.x && teleportLink.right.hex.y === coordinate.y && teleportLink.right.direction === direction) {
            return teleportLink.left.hex;
        }
    }
    return undefined;
}

export function applyMapDirection(map: GameMap, coordinate: Coordinate, direction: Direction): Coordinate {
    let teleportDest = applyTeleport(map, coordinate, direction);
    if (teleportDest !== undefined) {
        return teleportDest;
    }
    return applyDirection(coordinate, direction);
}

export function applyDirection(coordinate: Coordinate, direction: Direction): Coordinate {
    switch (direction) {
        case Direction.NORTH:
            return {x: coordinate.x, y: coordinate.y - 2}
        case Direction.NORTH_EAST:
            if ((coordinate.y % 2) === 0) {
                return {x: coordinate.x, y: coordinate.y - 1}
            } else {
                return {x: coordinate.x + 1, y: coordinate.y - 1}
            }
        case Direction.SOUTH_EAST:
            if ((coordinate.y % 2) === 0) {
                return {x: coordinate.x, y: coordinate.y + 1}
            } else {
                return {x: coordinate.x + 1, y: coordinate.y + 1}
            }
        case Direction.SOUTH:
            return {x: coordinate.x, y: coordinate.y + 2}
        case Direction.SOUTH_WEST:
            if ((coordinate.y % 2) === 0) {
                return {x: coordinate.x - 1, y: coordinate.y + 1}
            } else {
                return {x: coordinate.x, y: coordinate.y + 1}
            }
        case Direction.NORTH_WEST:
            if ((coordinate.y % 2) === 0) {
                return {x: coordinate.x - 1, y: coordinate.y - 1}
            } else {
                return {x: coordinate.x, y: coordinate.y - 1}
            }
    }
    throw new Error("unhandled direction: " + direction);
}

export function oppositeDirection(direction: Direction): Direction {
    switch (direction) {
        case Direction.NORTH: return Direction.SOUTH;
        case Direction.NORTH_EAST: return Direction.SOUTH_WEST;
        case Direction.SOUTH_EAST: return Direction.NORTH_WEST;
        case Direction.SOUTH: return Direction.NORTH;
        case Direction.SOUTH_WEST: return Direction.NORTH_EAST;
        case Direction.NORTH_WEST: return Direction.SOUTH_EAST;
    }
}

export function mapNameToDisplayName(mapName: string): string {
    if (mapName === 'rust_belt') {
        return "Rust Belt";
    }
    if (mapName === 'southern_us') {
        return "Southern U.S.";
    }
    if (mapName === 'germany') {
        return "Germany";
    }
    if (mapName === 'scotland') {
        return "Scotland";
    }
    return mapName;
}

export function specialActionToDisplayName(specialAction: SpecialAction): string {
    if (specialAction === 'first_move') {
        return "First Move";
    }
    if (specialAction === 'first_build') {
        return "First Build";
    }
    if (specialAction === 'engineer') {
        return "Engineer";
    }
    if (specialAction === 'loco') {
        return "Locomotive";
    }
    if (specialAction === 'urbanization') {
        return "Urbanization";
    }
    if (specialAction === 'production') {
        return "Production";
    }
    if (specialAction === 'turn_order_pass') {
        return "Turn Order Pass"
    }
    return specialAction;
}
