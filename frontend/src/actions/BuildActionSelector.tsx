import {
    BuildAction,
    ConfirmMove,
    Coordinate,
    Direction,
    GameState,
    Urbanization,
    User,
    ViewGameResponse
} from "../api/api.ts";
import {Button, Header, Icon} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {NewCitySelector,} from "./TrackSelector.tsx";
import ErrorContext from "../ErrorContext.tsx";
import maps, {BasicMap, HexType} from "../map.ts";
import {applyDirection, oppositeDirection} from "../util.ts";

function isCityHex(game: ViewGameResponse, map: BasicMap, urbanization: Urbanization|undefined, hex: Coordinate): boolean {
    if (map.hexes[hex.y][hex.x] === HexType.CITY) {
        return true;
    }
    for (let city of map.cities) {
        if (city.coordinate.x === hex.x && city.coordinate.y === hex.y) {
            return true;
        }
    }
    if (game.gameState && game.gameState.urbanizations) {
        for (let urb of game.gameState.urbanizations) {
            if (urb.hex.x === hex.x && urb.hex.y === hex.y) {
                return true;
            }
        }
    }
    if (urbanization) {
        if (urbanization.hex.x === hex.x && urbanization.hex.y === hex.y) {
            return true;
        }
    }

    return false;
}

function computeExistingRoutes(gameState: GameState|undefined, map: BasicMap): Array<Array<Array<{Left: Direction, Right: Direction}>>> {
    let routes: Array<Array<Array<{Left: Direction, Right: Direction}>>> = [];
    for (let y = 0; y < map.height; y++) {
        routes.push([]);
        for (let x = 0; x < map.width; x++) {
            routes[y].push([]);
        }
    }

    if (gameState && gameState.links) {
        for (let link of gameState.links) {
            let hex = link.sourceHex;
            for (let i = 1; i < link.steps.length; i++) {
                hex = applyDirection(hex, link.steps[i-1]);
                let left = oppositeDirection(link.steps[i-1]);
                let right = link.steps[i];
                routes[hex.y][hex.x].push({Left: left, Right: right});
            }
        }
    }
    return routes;
}

function BuildActionSelector({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [action, setAction] = useState<BuildAction>({
        townPlacements: [],
        trackRedirects: [],
        trackPlacements: [],
        urbanization: undefined,
    });
    let [showUrbanize, setShowUrbanize] = useState<boolean>(false);
    let [urbanizeSelection, setUrbanizeSelection] = useState<number>(0);
    let [buildingTrackHex, setBuildingTrackHex] = useState<Coordinate|undefined>(undefined);
    let [buildingTrackDirection, setBuildingTrackDirection] = useState<Direction|undefined>(undefined);
    let [loading, setLoading] = useState<boolean>(false);

    let map = maps[game.mapName];

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail) {
                if (showUrbanize) {
                    let newAction = Object.assign({}, action);
                    newAction.urbanization = {
                        city: urbanizeSelection,
                        hex: e.detail,
                    };
                    setShowUrbanize(false);
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                } else {
                    // Only change the hex if we weren't already building track
                    if (buildingTrackHex === undefined) {
                        setBuildingTrackHex(e.detail);
                        document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: e.detail }));
                    }
                }
            }
        };

        document.addEventListener('mapClickEvent', handler);
        return () => document.removeEventListener('mapClickEvent', handler);
    }, [action, showUrbanize, urbanizeSelection, buildingTrackHex]);

    useEffect(() => {
        const handler = (e:CustomEventInit<{direction: Direction}>) => {
            if (e.detail && buildingTrackHex !== undefined) {
                let direction = e.detail.direction;
                let priorDirection = buildingTrackDirection;
                let newHex = applyDirection(buildingTrackHex, direction);
                if (priorDirection !== undefined) {
                    if (map.hexes[buildingTrackHex.y][buildingTrackHex.x] === HexType.TOWN) {
                        // If we've built into a town, just ignore the direction and complete the link
                        let newAction = Object.assign({}, action);
                        newAction.townPlacements = newAction.townPlacements.slice();
                        newAction.townPlacements.push({
                            hex: buildingTrackHex,
                            track: oppositeDirection(priorDirection),
                        });
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));

                        // And clear the building step
                        setBuildingTrackHex(undefined);
                        setBuildingTrackDirection(undefined);
                        document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: undefined }));
                    } else {
                        // Just add some ordinary track...
                        let newAction = Object.assign({}, action);
                        newAction.trackPlacements = newAction.trackPlacements.slice();
                        newAction.trackPlacements.push({
                            hex: buildingTrackHex,
                            track: [oppositeDirection(priorDirection), direction],
                        });
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));

                        // If we've hit a city, clear the building step
                        if (isCityHex(game, map, newAction.urbanization, newHex)) {
                            setBuildingTrackHex(undefined);
                            setBuildingTrackDirection(undefined);
                            document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: undefined }));
                        } else {
                            // Otherwise keep going
                            setBuildingTrackHex(newHex);
                            setBuildingTrackDirection(direction);
                            document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: newHex }));
                        }
                    }
                } else {
                    let isCity = isCityHex(game, map, action.urbanization, buildingTrackHex);
                    // If building from a town (and it's not urbanized), add a town placement
                    if (map.hexes[buildingTrackHex.y][buildingTrackHex.x] === HexType.TOWN && !isCity) {
                        let newAction = Object.assign({}, action);
                        newAction.townPlacements = newAction.townPlacements.slice();
                        newAction.townPlacements.push({
                            hex: buildingTrackHex,
                            track: direction,
                        });
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                    } else if (isCity) {
                        // Do nothing, this is just defining where the route is coming from
                    } else {
                        // This is either redirect or extending existing track. We need to figure out what track already exists on this hex.
                        let existingRoutes = computeExistingRoutes(game.gameState, map)[buildingTrackHex.y][buildingTrackHex.x];
                        let isExistingRoute = false;
                        for (let route of existingRoutes) {
                            if (route.Left === direction || route.Right === direction) {
                                isExistingRoute = true;
                                break;
                            }
                        }
                        if (isExistingRoute) {
                            // Do nothing
                        } else {
                            // This is a redirect
                            let newAction = Object.assign({}, action);
                            newAction.trackRedirects = newAction.trackRedirects.slice();
                            newAction.trackRedirects.push({
                                hex: buildingTrackHex,
                                track: direction,
                            });
                            setAction(newAction);
                            document.dispatchEvent(new CustomEvent('pendingBuildAction', {detail: newAction}));
                        }
                    }
                    setBuildingTrackDirection(direction);
                    setBuildingTrackHex(newHex);
                    document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: newHex }));
                }
            }
        };

        document.addEventListener('arrowClickEvent', handler);
        return () => document.removeEventListener('arrowClickEvent', handler);
    }, [action, buildingTrackHex, buildingTrackDirection]);

    if (!game.gameState) {
        return null;
    }

    let playerById: { [playerId: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.activePlayer) {
        let activePlayer: User|undefined = playerById[game.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to build...</p>
    } else {
        if (showUrbanize) {
            content = <p>
                <p>Select new city to build, then click on hex:</p>
                <NewCitySelector selected={urbanizeSelection} onChange={(value) => {
                    setUrbanizeSelection(value)
                }} />
                <Button negative onClick={() => setShowUrbanize(false)}>Cancel</Button>
            </p>
        } else {
            let urbanizeButton: ReactNode = undefined;
            if (game.gameState.playerActions[game.activePlayer] === 'urbanization') {
                urbanizeButton = <>
                    <Button secondary disabled={!!action.urbanization} icon onClick={() => {
                        setBuildingTrackHex(undefined);
                        document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: undefined }));
                        setBuildingTrackDirection(undefined);
                        setUrbanizeSelection(0);
                        setShowUrbanize(true);
                    }}><Icon name="home" /> Urbanize</Button>
                </>
            }

            content = <>
                <p>To build track, select a starting hex (either a city, town, or a hex after the end of a dangling
                    track). Then click on the arrows to create track.</p>
                <p>If you build multiple track segments on a single tile, the builds will be consolidated (e.g. as a
                    complex track or a town with multiple legs) as a single tile placement when you submit the
                    action.</p>
                <p>To redirect track, select the hex with the dangling track and select a new direction to build (and
                    continue selecting directions to extend).</p>
                <p>To extend existing incomplete links, select the dangling track and continue in the direction of the
                    existing link.</p>
                <p>To leave a link unfinished, use the "unselect hex" button.</p>
                <div>
                    {urbanizeButton}
                    <Button primary loading={loading} onClick={() => {
                        setLoading(true);
                        ConfirmMove({
                            gameId: game.id,
                            actionName: "build",
                            buildAction: action,
                        }).then(() => {
                            let newAction: BuildAction = {
                                townPlacements: [],
                                trackRedirects: [],
                                trackPlacements: [],
                                urbanization: undefined,
                            };
                            setAction(newAction);
                            document.dispatchEvent(new CustomEvent('pendingBuildAction', {detail: newAction}));
                            setBuildingTrackHex(undefined);
                            setBuildingTrackDirection(undefined);
                            document.dispatchEvent(new CustomEvent('buildingTrackHex', {detail: undefined}));
                            return onDone();
                        }).catch(err => {
                            setError(err);
                        }).finally(() => {
                            setLoading(false);
                        });
                    }}>Finish Action</Button>
                    <Button negative loading={loading} onClick={() => {
                        let newAction: BuildAction = {
                            townPlacements: [],
                            trackRedirects: [],
                            trackPlacements: [],
                            urbanization: undefined,
                        };
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', {detail: newAction}));
                        setBuildingTrackHex(undefined);
                        setBuildingTrackDirection(undefined);
                        document.dispatchEvent(new CustomEvent('buildingTrackHex', {detail: undefined}));
                    }}>Restart Action</Button>
                    <Button basic disabled={buildingTrackHex === undefined} onClick={() => {
                        setBuildingTrackHex(undefined);
                        setBuildingTrackDirection(undefined);
                        document.dispatchEvent(new CustomEvent('buildingTrackHex', {detail: undefined}));
                    }}>Unselect Hex</Button>
                </div>
            </>;
        }
    }

    return <>
        <Header as='h2'>Building Phase</Header>
        {content}
    </>
}

export default BuildActionSelector
