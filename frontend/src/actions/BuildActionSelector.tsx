import {BuildAction, ConfirmMove, Coordinate, Direction, User, ViewGameResponse} from "../api/api.ts";
import {
    Button,
    Header,
    Icon,
    Modal,
    ModalActions,
    ModalContent,
    ModalDescription,
    ModalHeader,
    Segment
} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {TrackSelector,} from "./TrackSelector.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {HexType, maps} from "../maps";
import {NewCitySelector} from "./NewCitySelector.tsx";
import {renderHexCoordinate} from "../util.ts";

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
    let [loading, setLoading] = useState<boolean>(false);

    let map = maps[game.mapName];

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail) {
                if (showUrbanize) {
                    // Only urbanize on towns
                    if (map.getHexType(e.detail) != HexType.TOWN) {
                        return;
                    }
                    // Do not allow urbanization on top of existing urbanization
                    if (game.gameState && game.gameState.urbanizations) {
                        for (let urb of game.gameState.urbanizations) {
                            if (urb.hex.x === e.detail.x && urb.hex.y === e.detail.y) {
                                return;
                            }
                        }
                    }

                    let newAction = Object.assign({}, action);
                    newAction.urbanization = {
                        city: urbanizeSelection,
                        hex: e.detail,
                    };
                    setShowUrbanize(false);
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                } else {
                    setBuildingTrackHex(e.detail);
                }
            }
        };

        document.addEventListener('mapClickEvent', handler);
        return () => document.removeEventListener('mapClickEvent', handler);
    }, [action, showUrbanize, urbanizeSelection, buildingTrackHex]);

    useEffect(() => {
        const handler = (e:CustomEventInit<{hex: Coordinate, direction: Direction}>) => {
            if (e.detail) {
                let newAction = Object.assign({}, action);
                newAction.teleportLinkPlacements.push({
                    hex: e.detail.hex,
                    track: e.detail.direction,
                });
                setAction(newAction);
                document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
            }
        };

        document.addEventListener('teleportLinkClickEvent', handler);
        return () => document.removeEventListener('teleportLinkClickEvent', handler);
    }, [action, showUrbanize, urbanizeSelection, buildingTrackHex]);

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
                    return onDone();
                }).catch(err => {
                    setError(err);
                }).finally(() => {
                    setLoading(false);
                });
            };

            content = <>
                <p>To build track or upgrade a tile, click on a hex and select which track tile you want to place.</p>
                <p>To redirect track, select the hex with the dangling track and select a replacement track tile.</p>
                <div>
                    {urbanizeButton}
                    <Button primary loading={loading} onClick={() => {
                        if (!action.urbanization && urbanizeButton) {
                            setShowConfirmModal('urbanization');
                            return;
                        }
                        let buildLimit = map.getBuildLimit(game.gameState, game.activePlayer);
                        let cashOnHand = game.gameState?.playerCash[game.activePlayer];
                        if (cashOnHand !== undefined && cashOnHand >= 2 && action.townPlacements.length + action.trackPlacements.length + action.teleportLinkPlacements.length + action.trackRedirects.length < buildLimit) {
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
                    }}>Restart Action</Button>
                </div>
            </>;
        }
    }

    return <>
        <Header as='h2'>Building Phase</Header>
        {content}
        {buildingTrackHex === undefined ? null : <Segment>
                <Header as="h2">Building on hex {renderHexCoordinate(buildingTrackHex)}</Header>
                <TrackSelector coordinate={buildingTrackHex} map={map}
                               gameState={game.gameState} activePlayer={game.activePlayer}
                               onClick={(newTrackRoutes, newTownRoutes, redirectedRoute) => {
                    let newAction = Object.assign({}, action);
                    newAction.trackPlacements = newAction.trackPlacements.slice();
                    newAction.townPlacements = newAction.townPlacements.slice();
                    newAction.trackRedirects = newAction.trackRedirects.slice();
                    clearBuildsForHex(newAction, buildingTrackHex);
                    for (let newRoute of newTrackRoutes) {
                        newAction.trackPlacements.push({
                            track: newRoute,
                            hex: buildingTrackHex,
                        });
                    }
                    for (let newExit of newTownRoutes) {
                        newAction.townPlacements.push({
                            track: newExit,
                            hex: buildingTrackHex,
                        });
                    }
                    if (redirectedRoute !== undefined) {
                        newAction.trackRedirects.push({
                            track: redirectedRoute,
                            hex: buildingTrackHex,
                        });
                    }
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }} />
            <Button primary onClick={() => setBuildingTrackHex(undefined)}>OK</Button>
            <Button negative onClick={() => {
                let newAction = Object.assign({}, action);
                newAction.trackPlacements = newAction.trackPlacements.slice();
                newAction.townPlacements = newAction.townPlacements.slice();
                newAction.trackRedirects = newAction.trackRedirects.slice();
                clearBuildsForHex(newAction, buildingTrackHex);
                setAction(newAction);
                document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                setBuildingTrackHex(undefined);
            }}>Cancel</Button>
        </Segment>}
    </>
}

function clearBuildsForHex(action: BuildAction, hex: Coordinate) {
    for (let i = 0; i < action.trackPlacements.length; i++) {
        if (action.trackPlacements[i].hex.x === hex.x && action.trackPlacements[i].hex.y === hex.y) {
            action.trackPlacements.splice(i, 1);
            i -= 1;
        }
    }
    for (let i = 0; i < action.townPlacements.length; i++) {
        if (action.townPlacements[i].hex.x === hex.x && action.townPlacements[i].hex.y === hex.y) {
            action.townPlacements.splice(i, 1);
            i -= 1;
        }
    }
    for (let i = 0; i < action.trackRedirects.length; i++) {
        if (action.trackRedirects[i].hex.x === hex.x && action.trackRedirects[i].hex.y === hex.y) {
            action.trackRedirects.splice(i, 1);
            i -= 1;
        }
    }
}

export default BuildActionSelector
