import {Direction} from "../api/api.ts";
import {Grid, GridColumn, GridRow} from "semantic-ui-react";
import {ReactNode} from "react";
import {HexRenderer, urbCityState} from "./renderer/HexRenderer.tsx";
import {HexType} from "../map.ts";
import "./TrackSelector.css";

type TrackTile = Array<[Direction, Direction]>;

const TRACK_TILES: TrackTile[] = [
    // Simple
    // Straight
    [[Direction.NORTH, Direction.SOUTH]],
    // Gentle
    [[Direction.NORTH, Direction.SOUTH_EAST]],
    // Tight
    [[Direction.NORTH, Direction.NORTH_EAST]],

    // Complex crossing
    // X
    [[Direction.NORTH, Direction.SOUTH], [Direction.SOUTH_WEST, Direction.NORTH_EAST]],
    // Gentle X
    [[Direction.NORTH, Direction.SOUTH_EAST], [Direction.NORTH_EAST, Direction.SOUTH]],
    // Bow and arrow
    [[Direction.NORTH, Direction.SOUTH], [Direction.SOUTH_WEST, Direction.SOUTH_EAST]],

    // Complex co-existing
    // Baseball
    [[Direction.NORTH_WEST, Direction.NORTH_EAST], [Direction.SOUTH_WEST, Direction.SOUTH_EAST]],
    // Gentle+tight, left
    [[Direction.NORTH_WEST, Direction.SOUTH], [Direction.NORTH, Direction.NORTH_EAST]],
    // Gentle+tight, right
    [[Direction.NORTH_WEST, Direction.SOUTH], [Direction.NORTH_EAST, Direction.SOUTH_EAST]],
    // Straight and tight
    [[Direction.NORTH, Direction.SOUTH], [Direction.NORTH_EAST, Direction.SOUTH_EAST]],
    // Double tight
    [[Direction.NORTH, Direction.NORTH_EAST], [Direction.SOUTH_EAST, Direction.SOUTH]],
];

type TownTile = Array<Direction>;
const TOWN_TILES: TownTile[] = [
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

function rotateTrackTile(trackTile: TrackTile, n: number): TrackTile {
    let routes: Array<[Direction, Direction]> = [];
    for (let route of trackTile) {
        let rotatedRoute: [Direction, Direction] = [
            (route[0]+n)%6 as Direction, (route[1]+n)%6 as Direction
        ];
        routes.push(rotatedRoute);
    }
    return routes;
}

function rotateTownTile(townTile: TownTile, n: number): TownTile {
    let routes: Direction[] = [];
    for (let route of townTile) {
        let rotatedRoute: Direction = (route+n)%6
        routes.push(rotatedRoute);
    }
    return routes;
}

export function getTrackRoutesFromId(id: number): TrackTile {
    let trackTypeId = Math.floor(id/6);
    let rotation = id%6;
    return rotateTrackTile(TRACK_TILES[trackTypeId], rotation);
}

export function getTownRoutesFromId(id: number): TownTile {
    let trackTypeId = Math.floor(id/6);
    let rotation = id%6;
    return rotateTownTile(TOWN_TILES[trackTypeId], rotation);
}

interface Props {
    selected: number
    onChange: (selected: number) => void
}

export function TrackSelector(props: Props) {
    let rows: ReactNode[] = [];
    for (let trackTypeId = 0; trackTypeId < TRACK_TILES.length; trackTypeId++) {
        let trackTile = TRACK_TILES[trackTypeId];
        let columns: ReactNode[] = [];
        for (let rotation = 0; rotation < 6; rotation++) {
            let id = trackTypeId*6 + rotation;

            let rotatedTrack = rotateTrackTile(trackTile, rotation);
            let renderer = new HexRenderer(false);
            renderer.renderHex({x: 0, y: 0}, HexType.PLAINS);
            for (let route of rotatedTrack) {
                renderer.renderTrack({x: 0, y: 0}, route[0], route[1], undefined);
            }
            let classNames = "track-select";
            if (id === props.selected) {
                classNames += " selected";
            }

            columns.push(<GridColumn><div className={classNames} onClick={() => props.onChange(id)}>{renderer.render()}</div></GridColumn>)
        }
        rows.push(<GridRow columns="equal">{columns}</GridRow>);
    }

    return <Grid>
        {rows}
    </Grid>
}

export function TownTrackSelector(props: Props) {
    let rows: ReactNode[] = [];
    for (let trackTypeId = 0; trackTypeId < TOWN_TILES.length; trackTypeId++) {
        let trackTile = TOWN_TILES[trackTypeId];
        let columns: ReactNode[] = [];
        for (let rotation = 0; rotation < 6; rotation++) {
            let id = trackTypeId*6 + rotation;

            let rotatedTrack = rotateTownTile(trackTile, rotation);
            let renderer = new HexRenderer(false);
            renderer.renderHex({x: 0, y: 0}, HexType.TOWN);
            for (let route of rotatedTrack) {
                renderer.renderTownTrack({x: 0, y: 0}, route, undefined);
            }
            let classNames = "track-select";
            if (id === props.selected) {
                classNames += " selected";
            }

            columns.push(<GridColumn><div className={classNames} onClick={() => props.onChange(id)}>{renderer.render()}</div></GridColumn>)
        }
        rows.push(<GridRow columns="equal">{columns}</GridRow>);
    }

    return <Grid>
        {rows}
    </Grid>
}

export function NewCitySelector(props: Props) {
    let columns: ReactNode[] = [];
    for (let newCityNum = 0; newCityNum < 8; newCityNum++) {
        let renderer = new HexRenderer(false);
        renderer.renderCityHex({x: 0, y: 0}, urbCityState(newCityNum));

        let classNames = "track-select";
        if (newCityNum === props.selected) {
            classNames += " selected";
        }

        columns.push(<GridColumn><div className={classNames} onClick={() => props.onChange(newCityNum)}>{renderer.render()}</div></GridColumn>)
    }
    return <Grid>
        <GridRow columns="equal">
            {columns}
        </GridRow>
    </Grid>
}
