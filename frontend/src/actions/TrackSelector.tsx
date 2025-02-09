import {ReactNode, useContext} from "react";
import {HexRenderer} from "./renderer/HexRenderer.tsx";
import {GameMap} from "../maps";
import "./TrackSelector.css";
import {
    BasicTownTile,
    MapTileState,
    MapTrackTile,
    Rotation,
    Route,
    TOWN_TILES,
    TrackTile,
    TrackTiles,
} from "../game/map_state.ts";
import {Coordinate, Direction, GameState} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";

interface TrackSelectorProps {
    coordinate: Coordinate;
    map: GameMap;
    gameState: GameState;
    activePlayer: string;
    onClick: (newTrackPlacement: {tile: TrackTile, rotation: number}|undefined, newTownRoutes: Array<Direction>) => void
}

export function TrackSelector(props: TrackSelectorProps) {
    let userSession = useContext(UserSessionContext);
    let mapTileState = new MapTileState(props.map, props.gameState);

    let tracks: Array<ReactNode> = [];
    if (mapTileState.isTown(props.coordinate)) {
        for (let trackTypeId = 0; trackTypeId < TOWN_TILES.length; trackTypeId++) {
            let basicTownTile: BasicTownTile = TOWN_TILES[trackTypeId];
            for (let rotation: Rotation = 0; rotation < 6; rotation++) {
                if (!mapTileState.isValidTownPlacement(props.coordinate,
                    basicTownTile, rotation as Rotation)) {
                    continue;
                }

                let renderer = new HexRenderer(false, false, userSession);
                renderer.renderHex({x: 0, y: 0}, props.map.getHexType(props.coordinate), props.map.getLocationName(props.coordinate));
                for (let exit of mapTileState.getTileState(props.coordinate).routes) {
                    renderer.renderTownTrack({x: 0, y: 0}, exit.left, props.gameState.playerColor[exit.owner]);
                }
                let newExits = mapTileState.getNewTownExits(props.coordinate,
                    basicTownTile, rotation as Rotation);
                for (let newExit of newExits) {
                    renderer.renderTownTrack({x: 0, y: 0}, newExit, props.gameState.playerColor[props.activePlayer]);
                }
                let classNames = "track-select";

                tracks.push(<div className={classNames} onClick={() => {
                    props.onClick(undefined, newExits);
                }}>{renderer.render()}</div>);
            }
        }
    } else {
        for (let trackTile of TrackTiles) {
            for (let rotation: Rotation = 0; rotation < 6; rotation++) {
                if (!mapTileState.isValidTrackPlacement(props.coordinate,
                        trackTile, rotation as Rotation, props.activePlayer)) {
                    continue;
                }

                let renderer = new HexRenderer(false, false, userSession);
                renderer.renderHex({x: 0, y: 0}, props.map.getHexType(props.coordinate), props.map.getLocationName(props.coordinate));
                renderTrackTile(mapTileState.getTileState(props.coordinate).routes,
                    props.gameState, props.activePlayer, {x: 0, y: 0}, trackTile, rotation as Rotation, renderer);
                let classNames = "track-select";

                tracks.push(<div className={classNames} onClick={() => props.onClick({tile: trackTile, rotation: rotation}, [])}>{renderer.render()}</div>);
            }
        }
    }

    return <>
        <div style={{overflowX: "scroll", whiteSpace: "nowrap", paddingBottom: "1em"}}>
            {tracks}
        </div>
    </>
}

export function renderTrackTile(oldRoutes: Route[],
                                gameState: GameState,
                                activePlayer: string,
                                hex: Coordinate,
                                tile: TrackTile,
                                rotation: Rotation,
                                hexRenderer: HexRenderer) {
    let newRoutes = new MapTrackTile(tile, rotation).getRoutes();
    for (let newRoute of newRoutes) {
        let owner = activePlayer;
        for (let oldRoute of oldRoutes) {
            if (oldRoute.left === newRoute[0] || oldRoute.right === newRoute[0]
                || oldRoute.left === newRoute[1] || oldRoute.right === newRoute[1]) {
                owner = oldRoute.owner;
                break;
            }
        }
        hexRenderer.renderTrack(hex, newRoute[0], newRoute[1], gameState.playerColor[owner]);
    }
}
