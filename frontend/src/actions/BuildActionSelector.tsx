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
import {
    Button,
    Header,
    Icon,
    Modal,
    ModalActions,
    ModalContent,
    ModalDescription,
    ModalHeader
} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {NewCitySelector,} from "./TrackSelector.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {GameMap, HexType, maps} from "../maps";
import {applyMapDirection, applyTeleport, oppositeDirection} from "../util.ts";

function isCityHex(game: ViewGameResponse, map: GameMap, urbanization: Urbanization|undefined, hex: Coordinate): boolean {
    if (map.getHexType(hex) === HexType.CITY) {
        return true;
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

function computeExistingRoutes(gameState: GameState|undefined, map: GameMap): Array<Array<Array<{Left: Direction, Right: Direction}>>> {
    let routes: Array<Array<Array<{Left: Direction, Right: Direction}>>> = [];
    for (let y = 0; y < map.getHeight(); y++) {
        routes.push([]);
        for (let x = 0; x < map.getWidth(); x++) {
            routes[y].push([]);
        }
    }

    if (gameState && gameState.links) {
        for (let link of gameState.links) {
            let hex = link.sourceHex;
            for (let i = 1; i < link.steps.length; i++) {
                hex = applyMapDirection(map, hex, link.steps[i-1]);
                let left = oppositeDirection(link.steps[i-1]);
                let right = link.steps[i];
                routes[hex.y][hex.x].push({Left: left, Right: right});
            }
        }
    }
    return routes;
}

function ConfirmSkipBuildsModal({open, onConfirm, onCancel}: {open: 'urbanization'|'tracks'|undefined, onConfirm: () => void, onCancel: () => void}) {
    return (
        <Modal open={!!open}>
            <ModalHeader>Skip Actions?</ModalHeader>
            <ModalContent>
                <ModalDescription>
                    <Header>{open === 'urbanization' ? <>You haven't urbanized yet</> : <>You haven't built as much as you can</>}</Header>
                    <p>{open === 'urbanization' ? <>You haven't placed a new city. Do you really want to forego urbanization?</> : <>You haven't built on as many hexes as you can. Do you really want to forego builds?</>}</p>
                </ModalDescription>
            </ModalContent>
            <ModalActions>
                <Button primary onClick={onConfirm}>Yes, finish the action</Button>
                <Button negative onClick={onCancel}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

function BuildActionSelector({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [showConfirmModal, setShowConfirmModal] = useState<'urbanization'|'tracks'|undefined>(undefined);
    let [action, setAction] = useState<BuildAction>({
        townPlacements: [],
        trackRedirects: [],
        trackPlacements: [],
        urbanization: undefined,
        teleportLinkPlacements: []
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

                // If this was a teleport, we need to update the action
                let newAction = Object.assign({}, action);
                if (applyTeleport(map, buildingTrackHex, direction) !== undefined) {
                    newAction.teleportLinkPlacements = newAction.teleportLinkPlacements.slice();
                    newAction.teleportLinkPlacements.push({
                        hex: buildingTrackHex,
                        track: direction,
                    });
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', {detail: newAction}));
                }

                let newHex = applyMapDirection(map, buildingTrackHex, direction);
                if (priorDirection !== undefined) {
                    if (map.getHexType(buildingTrackHex) === HexType.TOWN) {
                        // If we've built into a town, just ignore the direction and complete the link
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
                    if (map.getHexType(buildingTrackHex) === HexType.TOWN && !isCity) {
                        newAction.townPlacements = newAction.townPlacements.slice();
                        newAction.townPlacements.push({
                            hex: buildingTrackHex,
                            track: direction,
                        });
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                    } else if (isCity) {
                        // If this is a city and the next is a city, check for an interurban link
                        let isNextCity = isCityHex(game, map, action.urbanization, newHex);
                        if (isNextCity) {
                            // Clear the selection
                            setBuildingTrackHex(undefined);
                            setBuildingTrackDirection(undefined);
                            document.dispatchEvent(new CustomEvent('buildingTrackHex', { detail: undefined }));
                            return;
                        } else {
                            // Do nothing, this is just defining where the route is coming from
                        }
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
            let alreadyUrbanized: number[] = [];

            if (game.gameState.urbanizations) {
                for (let urb of game.gameState.urbanizations) {
                    alreadyUrbanized.push(urb.city);
                }
            }

            content = <p>
                <p>Select new city to build, then click on hex:</p>
                <NewCitySelector selected={urbanizeSelection} alreadyUrbanized={alreadyUrbanized} onChange={(value) => {
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

            const commitAction = () => {
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
                        teleportLinkPlacements: [],
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
            };

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
                        if (!action.urbanization && urbanizeButton) {
                            setShowConfirmModal('urbanization');
                            return;
                        }
                        let buildLimit = map.getBuildLimit(game.gameState, game.activePlayer);
                        if (action.townPlacements.length + action.trackPlacements.length + action.teleportLinkPlacements.length + action.trackRedirects.length < buildLimit) {
                            setShowConfirmModal('tracks');
                            return;
                        }

                        commitAction();
                    }}>Finish Action</Button>
                    <ConfirmSkipBuildsModal
                        open={showConfirmModal}
                        onConfirm={() => {
                            setShowConfirmModal(undefined);
                            commitAction();
                        }}
                        onCancel={() => setShowConfirmModal(undefined)} />
                    <Button negative loading={loading} onClick={() => {
                        let newAction: BuildAction = {
                            townPlacements: [],
                            trackRedirects: [],
                            trackPlacements: [],
                            teleportLinkPlacements: [],
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
