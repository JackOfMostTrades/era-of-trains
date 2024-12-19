import {BuildAction, ConfirmMove, Coordinate, User, ViewGameResponse} from "../api/api.ts";
import {Button, Header, Icon} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {
    getTownRoutesFromId,
    getTrackRoutesFromId,
    NewCitySelector,
    TownTrackSelector,
    TrackSelector
} from "./TrackSelector.tsx";

interface Step {
    kind?: 'build_track' | 'build_town' | 'urbanize';
    buildTrackSelection: number;
    buildTownSelection: number;
    urbanizationSelection: number;
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
                if (step.kind === 'build_track') {
                    let newAction = Object.assign({}, action);
                    newAction.trackPlacements.push({
                        tracks: getTrackRoutesFromId(step.buildTrackSelection),
                        hex: e.detail,
                    });
                    setStep({});
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }
                if (step.kind === 'build_town') {
                    let newAction = Object.assign({}, action);
                    newAction.townPlacements.push({
                        tracks: getTownRoutesFromId(step.buildTownSelection),
                        hex: e.detail,
                    });
                    setStep({});
                    setAction(newAction);
                    document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                }
                if (step.kind === 'urbanize') {
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
                <TrackSelector selected={step.buildTrackSelection} onChange={(value) => {
                    setStep({kind: 'build_track', buildTrackSelection: value, buildTownSelection: 0, urbanizationSelection: 0});
                }} />
                <Button negative onClick={() => setStep({buildTrackSelection: 0, buildTownSelection: 0, urbanizationSelection: 0})}>Cancel</Button>
            </>
        } else if (step.kind === 'build_town') {
            content = <p>
                <p>Select town to build, then click on hex:</p>
                <TownTrackSelector selected={step.buildTownSelection} onChange={(value) => {
                    setStep({
                        kind: 'build_town',
                        buildTrackSelection: 0,
                        buildTownSelection: value,
                        urbanizationSelection: 0
                    });
                }}/>
                <Button negative onClick={() => setStep({
                    buildTrackSelection: 0,
                    buildTownSelection: 0,
                    urbanizationSelection: 0
                })}>Cancel</Button>
            </p>
        } else if (step.kind === 'urbanize') {
            content = <p>
                <p>Select new city to build, then click on hex:</p>
                <NewCitySelector selected={step.urbanizationSelection} onChange={(value) => {
                    setStep({
                        kind: 'urbanize',
                        buildTrackSelection: 0,
                        buildTownSelection: 0,
                        urbanizationSelection: value
                    });
                }}/>
                <Button negative onClick={() => setStep({
                    buildTrackSelection: 0,
                    buildTownSelection: 0,
                    urbanizationSelection: 0
                })}>Cancel</Button>
            </p>
        } else {
            let urbanizeButton: ReactNode = undefined;
            if (game.gameState.playerActions[game.gameState.activePlayer] === 'urbanization') {
                urbanizeButton = <>
                    <Button secondary disabled={!!action.urbanization} icon onClick={() => {
                        setStep({kind: 'urbanize', buildTrackSelection: 0, buildTownSelection: 0, urbanizationSelection: 0});
                    }}><Icon name="home" /> Urbanize</Button>
                </>
            }

            content = <>
                <p>Select build step:</p>
                <div>
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
                        let newAction: BuildAction = {
                            townPlacements: [],
                            trackPlacements: [],
                            urbanization: undefined,
                        };
                        setAction(newAction);
                        document.dispatchEvent(new CustomEvent('pendingBuildAction', { detail: newAction }));
                    }}>Restart Action</Button>
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
