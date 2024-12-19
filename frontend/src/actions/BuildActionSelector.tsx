import {BuildAction, ConfirmMove, Coordinate, Direction, User, ViewGameResponse} from "../api/api.ts";
import {Button, ButtonGroup, Dropdown, DropdownItemProps, Header, Icon} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";

interface Step {
    kind?: 'build_track' | 'build_town' | 'urbanize';
    buildTrackSelection?: number;
    buildTownSelection?: number;
    urbanizationSelection?: number;
}

const TRACK_OPTIONS: DropdownItemProps[] = [
    {
        key: 0,
        value: 0,
        text: "Simple straight: north/south"
    },
    {
        key: 1,
        value: 1,
        text: "Simple straight: northeast/southwest"
    },
    {
        key: 2,
        value: 2,
        text: "Simple straight: northwest/southeast"
    },
    {
        key: 3,
        value: 3,
        text: "Gentle: north/southeast"
    }
];
function trackOptionToDirections(value: number): Array<[Direction, Direction]> {
    switch (value) {
        case 0: return [[Direction.SOUTH, Direction.NORTH]]
        case 1: return [[Direction.NORTH_EAST, Direction.SOUTH_WEST]]
        case 2: return [[Direction.NORTH_WEST, Direction.SOUTH_EAST]]
        case 3: return [[Direction.NORTH, Direction.SOUTH_EAST]]
    }
    throw new Error("Unhandled value: " + value);
}

function townOptionToDirections(value: number): Array<Direction> {
    throw new Error("Unhandled value: " + value);
}

function BuildActionSelector({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [action, setAction] = useState<BuildAction>({
        townPlacements: [],
        trackPlacements: [],
        urbanization: undefined,
    });
    let [step, setStep] = useState<Step>({})
    let [loading, setLoading] = useState<boolean>(false);

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail) {
                if (step.kind === 'build_track' && step.buildTrackSelection !== undefined) {
                    let newAction = Object.assign({}, action);
                    newAction.trackPlacements.push({
                        tracks: trackOptionToDirections(step.buildTrackSelection),
                        hex: e.detail,
                    });
                    setStep({});
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }
                if (step.kind === 'build_town' && step.buildTownSelection) {
                    let newAction = Object.assign({}, action);
                    newAction.townPlacements.push({
                        tracks: townOptionToDirections(step.buildTownSelection),
                        hex: e.detail,
                    });
                    setStep({});
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }
                if (step.kind === 'urbanize' && step.urbanizationSelection !== undefined) {
                    let newAction = Object.assign({}, action);
                    newAction.urbanization = {
                        city: step.urbanizationSelection,
                        hex: e.detail,
                    };
                    setStep({});
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }
            }
        };

        document.addEventListener('mapClickEvent', handler);
        return () => document.removeEventListener('mapClickEvent', handler);
    }, [action, step]);

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
        content = <p>Waiting for {activePlayer?.nickname} to build...</p>
    } else {
        if (step.kind === 'build_track') {
            content = <>
                <p>Select track to build, then click on hex:</p>
                <Dropdown selection
                          value={step.buildTrackSelection}
                          options={TRACK_OPTIONS}
                          onChange={(_, { value }) => {
                              setStep({kind: 'build_track', buildTrackSelection: value as number});
                          }} />
                <Button negative onClick={() => setStep({})}>Cancel</Button>
            </>
        } else if (step.kind === 'build_town') {
            content = <p>
                <Button negative onClick={() => setStep({})}>Cancel</Button>
            </p>
        } else if (step.kind === 'urbanize') {
            content = <p>
                <Button negative onClick={() => setStep({})}>Cancel</Button>
            </p>
        } else {
            let urbanizeButton: ReactNode = undefined;
            if (game.gameState.playerActions[game.gameState.activePlayer] === 'urbanization') {
                urbanizeButton = <>
                    <Button secondary disabled={!!action.urbanization}>Urbanize</Button>
                </>
            }

            content = <>
                <p>Select build step:</p>
                <ButtonGroup>
                    <Button secondary icon onClick={() => {
                        setStep({kind: 'build_track', buildTrackSelection: 0});
                    }}><Icon name='train' /> Build Track</Button>
                    <Button secondary icon onClick={() => {
                        setStep({kind: 'build_town', buildTownSelection: 0})
                    }}><Icon name='circle' /> Build Town</Button>
                    {urbanizeButton}
                    <Button primary loading={loading} onClick={() => {
                        setLoading(true);
                        ConfirmMove({
                            gameId: game.id,
                            actionName: "build",
                            buildAction: action,
                        }).then(() => {
                            return onDone();
                        }).finally(() => {
                            setLoading(false);
                        });
                    }}>Finish Action</Button>
                    <Button negative loading={loading} onClick={() => {
                        setAction({
                            townPlacements: [],
                            trackPlacements: [],
                            urbanization: undefined,
                        });
                    }}>Restart Action</Button>
                </ButtonGroup>
            </>;
        }
    }

    return <>
        <Header as='h2'>Building Phase</Header>
        {content}
    </>
}

export default BuildActionSelector
