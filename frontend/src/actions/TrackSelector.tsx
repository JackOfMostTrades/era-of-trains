import {ReactNode, useContext} from "react";
import {HexRenderer} from "./renderer/HexRenderer.tsx";
import {GameMap} from "../maps";
import "./TrackSelector.css";
import {BasicTownTile, MapTileState, Rotation, TOWN_TILES, TrackTiles,} from "../game/map_state.ts";
import {Coordinate, Direction, GameState} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";

interface TrackSelectorProps {
    coordinate: Coordinate;
    map: GameMap;
    gameState: GameState;
    activePlayer: string;
    onClick: (newTrackRoutes: Array<[Direction, Direction]>, newTownRoutes: Array<Direction>, redirectedRoute: Direction|undefined) => void
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
                renderer.renderHex({x: 0, y: 0}, props.map.getHexType(props.coordinate));
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
                    props.onClick([], newExits, undefined);
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
                let redirectedRoutes = mapTileState.getRedirectedRoutes(props.coordinate, trackTile, rotation as Rotation);
                // FIXME: Backend only supports redirecting a single track at a time
                if (redirectedRoutes.length > 1) {
                    continue;
                }

                let renderer = new HexRenderer(false, false, userSession);
                renderer.renderHex({x: 0, y: 0}, props.map.getHexType(props.coordinate));
                for (let route of mapTileState.getTileState(props.coordinate).routes) {
                    let skip = false;
                    if (route.owner === "" || route.owner === props.activePlayer) {
                        let isDangler = false;
                        for (let dangler of mapTileState.getTileState(props.coordinate).danglers) {
                            if (dangler.from === route.left || dangler.from === route.right) {
                                isDangler = true;
                                break;
                            }
                        }
                        if (isDangler) {
                            skip = true;
                        }
                    }
                    if (!skip) {
                        renderer.renderTrack({
                            x: 0,
                            y: 0
                        }, route.left, route.right, props.gameState.playerColor[route.owner]);
                    }
                }
                let newRoutes = mapTileState.getNewRoutes(props.coordinate, trackTile, rotation as Rotation);
                for (let route of newRoutes) {
                    renderer.renderTrack({x: 0, y: 0}, route[0], route[1], props.gameState.playerColor[props.activePlayer]);
                }
                for (let route of redirectedRoutes) {
                    renderer.renderTrack({x: 0, y: 0}, route[0], route[1], props.gameState.playerColor[props.activePlayer]);
                }
                let classNames = "track-select";

                tracks.push(<div className={classNames} onClick={() => props.onClick(newRoutes, [],
                    redirectedRoutes.length === 0 ? undefined : redirectedRoutes[0][1])}>{renderer.render()}</div>);
            }
        }
    }

    return <>
        <div style={{overflowX: "scroll", whiteSpace: "nowrap", paddingBottom: "1em"}}>
            {tracks}
        </div>
    </>
}
