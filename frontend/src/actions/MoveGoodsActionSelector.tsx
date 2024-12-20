import {Color, ConfirmMove, Coordinate, Direction, User, ViewGameResponse} from "../api/api.ts";
import {Button, Header, Icon} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {applyDirection, oppositeDirection} from "../util.ts";
import ErrorContext from "../ErrorContext.tsx";

export interface Step {
    selectedColor: Color;
    selectedOrigin: Coordinate;
    steps: Direction[];
    currentCubePosition: Coordinate;
}

function getDirectionFromSource(source: Coordinate, target: Coordinate): Direction|undefined {
    for (let direction of [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH_EAST, Direction.SOUTH, Direction.SOUTH_WEST, Direction.NORTH_WEST]) {
        let step = applyDirection(source, direction);
        if (step.x === target.x && step.y === target.y) {
            return direction;
        }
    }
    return undefined;
}

function computeNextStop(game: ViewGameResponse, current: Coordinate, direction: Direction): Coordinate|undefined {
    for (let link of game.gameState?.links) {
        if (!link.complete) {
            continue;
        }
        let end = link.sourceHex;
        for (let dir of link.steps) {
            end = applyDirection(end, dir);
        }
        if (link.sourceHex.x === current.x && link.sourceHex.y === current.y && link.steps[0] === direction) {
            return end;
        }
        if (end.x === current.x && end.y === current.y && direction === oppositeDirection(link.steps[link.steps.length-1])) {
            return link.sourceHex;
        }
    }
    return undefined;
}

function MoveGoodsActionSelector({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [step, setStep] = useState<Step>({selectedColor: Color.NONE, selectedOrigin: {x: 0, y: 0}, steps: [], currentCubePosition: {x: 0, y: 0}})
    let [loading, setLoading] = useState<boolean>(false);

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail && step.selectedColor !== Color.NONE) {
                let direction = getDirectionFromSource(step.currentCubePosition, e.detail);
                if (direction !== undefined) {
                    let nextStop = computeNextStop(game, step.currentCubePosition, direction);
                    if (nextStop !== undefined) {
                        let newStep = Object.assign({}, step);
                        newStep.steps = newStep.steps.slice();
                        newStep.steps.push(direction);
                        newStep.currentCubePosition = nextStop;
                        setStep(newStep);
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                    }
                }
            }
        };

        document.addEventListener('mapClickEvent', handler);
        return () => document.removeEventListener('mapClickEvent', handler);
    }, [step]);

    useEffect(() => {
        const handler = (e:CustomEventInit<{x:number, y:number, color: Color}>) => {
            if (e.detail && e.detail.color !== Color.NONE) {
                let newStep = {
                    selectedColor: e.detail.color,
                    selectedOrigin: {x: e.detail.x, y: e.detail.y},
                    steps: [],
                    currentCubePosition: {x: e.detail.x, y: e.detail.y},
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

    if (userSession.userInfo?.user.id !== game.gameState.activePlayer) {
        let activePlayer: User|undefined = playerById[game.gameState.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to move goods...</p>
    } else {
        let hasDoneLoco = false;
        if (game.gameState.playerHasDoneLoco[game.gameState.activePlayer]) {
            hasDoneLoco = true;
        }

        content = <>
            <p>Select move goods action:<br/>To move a good, select the cube on the map, then click on the neighboring hex in the direction you want to move it. Press the finish button when the cube is at its destination.</p>
            <div>
                <Button disabled={step.selectedColor === Color.NONE || step.steps.length === 0} primary icon onClick={() => {
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
                        let newStep = {selectedColor: Color.NONE, selectedOrigin: {x: 0, y: 0}, steps: [], currentCubePosition: {x: 0, y: 0}};
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}><Icon name='square' /> Finish moving good</Button>
                <Button disabled={hasDoneLoco} secondary icon onClick={() => {
                    setLoading(true);
                    ConfirmMove({
                        gameId: game.id,
                        actionName: "move_goods",
                        moveGoodsAction: {loco: true},
                    }).then(() => {
                        let newStep = {selectedColor: Color.NONE, selectedOrigin: {x: 0, y: 0}, steps: [], currentCubePosition: {x: 0, y: 0}};
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}><Icon name='train' /> Increase Locomotive</Button>
                <Button negative loading={loading} onClick={() => {
                    setLoading(true);
                    ConfirmMove({
                        gameId: game.id,
                        actionName: "move_goods",
                        moveGoodsAction: {},
                    }).then(() => {
                        let newStep = {selectedColor: Color.NONE, selectedOrigin: {x: 0, y: 0}, steps: [], currentCubePosition: {x: 0, y: 0}};
                        document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                        return onDone();
                    }).catch(err => {
                        setError(err);
                    }).finally(() => {
                        setLoading(false);
                    });
                }}>Pass</Button>
                <Button negative secondary loading={loading} onClick={() => {
                    let newStep: Step = {
                        selectedColor: Color.NONE,
                        selectedOrigin: {x: 0, y: 0},
                        steps: [],
                        currentCubePosition: {x: 0, y: 0}
                    };
                    setStep(newStep);
                    document.dispatchEvent(new CustomEvent('pendingMoveGoods', { detail: newStep }));
                }}>Restart Action</Button>
            </div>
        </>;
    }

    return <>
        <Header as='h2'>Move Goods</Header>
        {content}
    </>
}

export default MoveGoodsActionSelector
