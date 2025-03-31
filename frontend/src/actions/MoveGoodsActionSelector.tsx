import {
    ALL_DIRECTIONS,
    Color,
    ConfirmMove,
    Coordinate,
    Direction,
    PlayerColor,
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
import {applyDirection, applyTeleport, oppositeDirection} from "../util.ts";
import ErrorContext from "../ErrorContext.tsx";
import {GameMap, maps} from "../maps";

export interface Step {
    selectedColor?: Color;
    selectedOrigin?: Coordinate;
    steps?: Direction[];
    moveAlong?: Coordinate[];
    currentCubePosition?: Coordinate;
    nextStepOptions?: Array<{direction: Direction, owner: PlayerColor|undefined}>
    playerToLinkCount?: { [ playerId: string]: number }
}

function ConfirmPassModal({open, hasLocoOption, onConfirm, onCancel}: {open: boolean, hasLocoOption: boolean, onConfirm: () => void, onCancel: () => void}) {
    return (
        <Modal open={open}>
            <ModalHeader>Pass your move goods actions?</ModalHeader>
            <ModalContent>
                <ModalDescription>
                    <Header>Do you really want to pass?</Header>
                    <p>Are you sure you want to pass instead of moving a good?</p>
                    {hasLocoOption && <p>You also still have the option to increase locomotive instead!</p>}
                </ModalDescription>
            </ModalContent>
            <ModalActions>
                <Button primary onClick={onConfirm}>Yes, skip doing anything</Button>
                <Button negative onClick={onCancel}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

function computeNextStop(game: ViewGameResponse, map: GameMap, current: Coordinate, direction: Direction): { end: Coordinate, moveAlong: Coordinate[], linkOwner: string }|undefined {
    if (!game.gameState || !game.gameState.links) {
        return undefined;
    }
    for (let link of game.gameState.links) {
        if (!link.complete) {
            continue;
        }
        let end = link.sourceHex;
        let lastReverseDirection: Direction = Direction.NORTH;
        let moveAlong = [end];
        for (let dir of link.steps) {
            let teleportDest = applyTeleport(map, game.gameState, undefined, end, dir);
            if (teleportDest !== undefined) {
                end = teleportDest.hex;
                lastReverseDirection = teleportDest.direction;
            } else {
                end = applyDirection(end, dir);
                lastReverseDirection = oppositeDirection(dir);
            }
            moveAlong.push(end);
        }
        if (link.sourceHex.x === current.x && link.sourceHex.y === current.y && link.steps[0] === direction) {
            moveAlong = moveAlong.slice(0, moveAlong.length-1);
            return {end: end, moveAlong: moveAlong, linkOwner: link.owner};
        }
        if (end.x === current.x && end.y === current.y && direction === lastReverseDirection) {
            moveAlong = moveAlong.reverse().slice(0, moveAlong.length-1);
            return {end: link.sourceHex, moveAlong: moveAlong, linkOwner: link.owner};
        }
    }
    return undefined;
}

function getValidStepDirections(game: ViewGameResponse, map: GameMap, origin: Coordinate): Array<{direction: Direction, owner: string}> {
    let results = [];
    for (let direction of ALL_DIRECTIONS) {
        let nextStop = computeNextStop(game, map, origin, direction);
        if (nextStop !== undefined) {
            results.push({direction: direction, owner: nextStop.linkOwner});
        }
    }
    return results;
}

function MoveGoodsActionSelector({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [step, setStep] = useState<Step>({})
    let [loading, setLoading] = useState<boolean>(false);
    let [showConfirmPassModal, setShowConfirmPassModal] = useState<boolean>(false);

    let map = maps[game.mapName];

    useEffect(() => {
        const handler = (e:CustomEventInit<{direction: Direction}>) => {
            if (e.detail && step.selectedColor !== undefined && step.currentCubePosition) {
                let direction = e.detail.direction;
                let nextStop = computeNextStop(game, map, step.currentCubePosition, direction);
                if (nextStop !== undefined) {
                    let newStep = Object.assign({}, step);
                    newStep.steps = (newStep.steps || []).slice();
                    newStep.steps.push(direction);
                    newStep.playerToLinkCount = Object.assign({}, newStep.playerToLinkCount);
                    if (newStep.playerToLinkCount[nextStop.linkOwner]) {
                        newStep.playerToLinkCount[nextStop.linkOwner] += 1;
                    } else {
                        newStep.playerToLinkCount[nextStop.linkOwner] = 1;
                    }
                    newStep.moveAlong = nextStop.moveAlong;
                    newStep.currentCubePosition = nextStop.end;
                    newStep.nextStepOptions = [];
                    for (let option of getValidStepDirections(game, map, nextStop.end)) {
                        newStep.nextStepOptions?.push({direction: option.direction, owner: game.gameState?.playerColor[option.owner]});
                    }
                    setStep(newStep);
                    document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                }
            }
        };

        document.addEventListener('arrowClickEvent', handler);
        return () => document.removeEventListener('arrowClickEvent', handler);
    }, [step]);

    useEffect(() => {
        const handler = (e:CustomEventInit<{x:number, y:number, color: Color}>) => {
            if (e.detail && e.detail.color !== Color.NONE) {
                let hex: Coordinate = {x: e.detail.x, y: e.detail.y};
                let newStep: Step = {
                    selectedColor: e.detail.color,
                    selectedOrigin: hex,
                    steps: [],
                    moveAlong: undefined,
                    currentCubePosition: hex,
                    nextStepOptions: [],
                    playerToLinkCount: {}
                }
                for (let option of getValidStepDirections(game, map, hex)) {
                    newStep.nextStepOptions?.push({direction: option.direction, owner: game.gameState?.playerColor[option.owner]});
                }

                setStep(newStep);
                document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
            }
        };

        document.addEventListener('cubeClickEvent', handler);
        return () => document.removeEventListener('cubeClickEvent', handler);
    }, [step]);

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
        content = <p>Waiting for {activePlayer?.nickname} to move goods...</p>
    } else {
        let hasLocoOption = true;
        if (game.gameState.playerHasDoneLoco[game.activePlayer] || game.gameState.playerLoco[game.activePlayer] >= 6) {
            hasLocoOption = false
        }
        let stepCountLabel: ReactNode = null;
        if (step.playerToLinkCount && step.steps && step.steps.length > 0) {
            let myCount = step.playerToLinkCount[game.activePlayer] || 0;
            if (myCount === step.steps.length) {
                stepCountLabel = " (" + myCount + ")";
            } else {
                stepCountLabel = " (" + myCount + " for me, " + (step.steps.length-myCount) + " for others)";
            }
        }

        content = <>
            <ConfirmPassModal
                open={showConfirmPassModal}
                hasLocoOption={hasLocoOption}
                onCancel={() => setShowConfirmPassModal(false)}
                onConfirm={() => {
                    setLoading(true);
                    setShowConfirmPassModal(false);
                    ConfirmMove({
                        gameId: game.id,
                        actionName: "move_goods",
                        moveGoodsAction: {},
                    }).then(() => {
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: {} }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}/>
            <p>Select move goods action:<br/>To move a good, select the cube on the map, then click on one of the arrows that appears to indicate the link you want to move it along. Press the finish button when the cube is at its final destination.</p>
            <div>
                <Button disabled={step.selectedColor === Color.NONE || !step.steps || step.steps.length === 0} primary icon onClick={() => {
                    setLoading(true);
                    ConfirmMove({
                        gameId: game.id,
                        actionName: "move_goods",
                        moveGoodsAction: {
                            startingLocation: step.selectedOrigin,
                            color: step.selectedColor,
                            path: step.steps,
                        },
                    }).then(() => {
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: {} }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}><Icon name='square' /> Finish moving good{stepCountLabel}</Button>
                <Button disabled={!hasLocoOption} secondary icon onClick={() => {
                    setLoading(true);
                    ConfirmMove({
                        gameId: game.id,
                        actionName: "move_goods",
                        moveGoodsAction: {loco: true},
                    }).then(() => {
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: {} }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}><Icon name='train' /> Increase Locomotive</Button>
                <Button negative loading={loading} onClick={() => setShowConfirmPassModal(true)}>Pass</Button>
                <Button negative secondary loading={loading} onClick={() => {
                    let newStep: Step = {
                        selectedColor: Color.NONE,
                        selectedOrigin: {x: 0, y: 0},
                        steps: [],
                        moveAlong: undefined,
                        currentCubePosition: {x: 0, y: 0},
                        playerToLinkCount: {}
                    };
                    setStep(newStep);
                    document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                }}>Restart Action</Button>
            </div>
        </>;
    }

    return <>
        <Header as='h2'>Move Goods (Round {game.gameState.movingGoodsRound+1})</Header>
        {content}
    </>
}

export default MoveGoodsActionSelector
