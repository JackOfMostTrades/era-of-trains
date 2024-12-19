import {Coordinate, Direction} from "./api/api.ts";

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