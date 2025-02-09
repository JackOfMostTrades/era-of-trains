import {Button, Grid, Header, Label, LabelDetail, List, ListItem, Loader, Segment} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import {useParams} from "react-router";
import {
    GamePhase,
    GetGameLogs,
    GetGameLogsResponse,
    JoinGame,
    LeaveGame,
    PlayerColor,
    PollGameStatus,
    StartGame,
    UndoMove,
    User,
    ViewGame,
    ViewGameResponse
} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";
import ChooseShares from "../actions/ChooseShares.tsx";
import AuctionAction from "../actions/AuctionAction.tsx";
import SpecialActionChooser from "../actions/SpecialActionChooser.tsx";
import BuildActionSelector from "../actions/BuildActionSelector.tsx";
import ViewMapComponent from "./ViewMapComponent.tsx";
import MoveGoodsActionSelector from "../actions/MoveGoodsActionSelector.tsx";
import GoodsGrowthTable from "./GoodsGrowthTable.tsx";
import ProductionAction from "../actions/ProductionAction.tsx";
import {playerColorToHtml} from "../actions/renderer/HexRenderer.tsx";
import GameLogsComponent from "./GameLogsComponent.tsx";
import FinalScore from "../actions/FinalScore.tsx";
import {GameMap, maps} from "../maps";
import "./ViewGamePage.css";
import {mapNameToDisplayName, specialActionToDisplayName} from "../util.ts";
import GameChat from "../components/GameChat.tsx";
import ErrorContext from "../ErrorContext.tsx";
import PartsCountComponent from "./PartsCountComponent.tsx";

function WaitingForPlayersPage({game, onJoin, onStart}: {game: ViewGameResponse, onJoin: () => Promise<void>, onStart: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [loading, setLoading] = useState<boolean>(false);

    let joined = false;
    for (let player of game?.joinedUsers) {
        if (player.id === userSession.userInfo?.user.id) {
            joined = true;
            break;
        }
    }

    let listItems: ReactNode[] = [];
    for (let i = 0; i < game.joinedUsers.length; i++) {
        let player = game.joinedUsers[i];
        listItems.push(<ListItem>{player.nickname}</ListItem>);
    }

    const isOwner = userSession.userInfo?.user.id === game.ownerUser.id;
    const map = maps[game.mapName];

    return <>
        <p>Waiting for players...</p>
        <List>{listItems}</List>
        {isOwner ? null :
            joined ? <>
                    <Button negative loading={loading} onClick={() => {
                        setLoading(true)
                        LeaveGame({gameId: game.id}).then(() => {
                            return onJoin();
                        }).finally(() => {
                            setLoading(false);
                        })
                    }}>Leave Game</Button>
                </> : <>
                    <Button primary loading={loading} disabled={game.joinedUsers.length >= game.maxPlayers} onClick={() => {
                        setLoading(true)
                        JoinGame({gameId: game.id}).then(() => {
                            return onJoin();
                        }).finally(() => {
                            setLoading(false);
                        })
                    }}>Join Game</Button>
                </>
        }
        {!isOwner || game?.joinedUsers.length < game?.minPlayers ? null : <>
            <Button primary loading={loading} disabled={userSession.userInfo?.user.id !== game.ownerUser.id} onClick={() => {
                setLoading(true)
                StartGame({gameId: game.id}).then(() => {
                    return onStart();
                }).finally(() => {
                    setLoading(false);
                })
            }}>Start Game</Button>
        </>}

        <Segment>
            <Header as="h2">{mapNameToDisplayName(game.mapName)}</Header>
            {map.getMapInfo()}
            <ViewMapComponent gameState={undefined} activePlayer="" map={map} />
        </Segment>
    </>
}

function PlayerColorAndName({nickname, color}: {nickname: string, color: PlayerColor|undefined}) {
    let userSession = useContext(UserSessionContext);

    return <><div style={{
        height: '1em',
        width: '1em',
        borderRadius: '50%',
        display: 'inline-block',
        backgroundColor: playerColorToHtml(color, userSession)
    }}/> {nickname}</>
}

function PlayerStatus({game, map, onConfirmMove}: { game: ViewGameResponse, map: GameMap, onConfirmMove: () => Promise<void> }) {
    let userSession = useContext(UserSessionContext);

    if (!game.gameState) {
        return null;
    }

    let playerById: { [id: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let playerInfoOrder: string[] = [];
    for (let player of game.joinedUsers) {
        playerInfoOrder.push(player.id);
    }
    playerInfoOrder.sort((a,b) => {
        return (game.gameState?.playerColor[a] || 0) - (game.gameState?.playerColor[b] || 0);
    });

    let playerColumns: ReactNode[] = [];
    for (let playerId of playerInfoOrder) {
        let nickname = playerById[playerId].nickname;
        let deltaLabel: string|null;
        if (game.gameState.playerIncome[playerId] >= 0) {
            let delta = game.gameState.playerIncome[playerId] - game.gameState.playerLoco[playerId] - game.gameState.playerShares[playerId];
            if (delta >= 0) {
                deltaLabel = "(+$" + delta + ")";
            } else {
                deltaLabel = "(-$" + Math.abs(delta) + ")";
            }
        } else {
            deltaLabel = null;
        }

        playerColumns.push(<Grid.Column key={playerId}><Segment>
                Player: <PlayerColorAndName nickname={nickname} color={game.gameState.playerColor[playerId]} /><br/>
                Cash: ${game.gameState.playerCash[playerId]} {deltaLabel}<br/>
                Shares: {game.gameState.playerShares[playerId]}<br/>
                Income: {game.gameState.playerIncome[playerId]}<br/>
                Loco: {game.gameState.playerLoco[playerId]}<br/>
                Special Action: {specialActionToDisplayName(game.gameState.playerActions[playerId])}<br/>
            </Segment></Grid.Column>);
    }

    let actionHolder: ReactNode;
    if (game.finished) {
        actionHolder = <FinalScore game={game} />
    } else {
        switch (game.gameState.gamePhase) {
            case GamePhase.SHARES:
                actionHolder = <ChooseShares game={game} map={map} onDone={onConfirmMove}/>
                break;
            case GamePhase.AUCTION:
                actionHolder = <AuctionAction game={game} onDone={onConfirmMove}/>
                break;
            case GamePhase.CHOOSE_SPECIAL_ACTIONS:
                actionHolder = <SpecialActionChooser game={game} onDone={onConfirmMove}/>
                break;
            case GamePhase.BUILDING:
                actionHolder = <BuildActionSelector game={game} onDone={onConfirmMove}/>
                break;
            case GamePhase.MOVING_GOODS:
                actionHolder = <MoveGoodsActionSelector game={game} onDone={onConfirmMove}/>
                break;
            case GamePhase.GOODS_GROWTH:
                actionHolder = <ProductionAction game={game} onDone={onConfirmMove}/>
                break;
        }
    }

    return <>
        <Segment>
            <Header as='h2'>Game Status</Header>
            <Grid columns={4} doubling stackable>
                {playerColumns}
            </Grid>
            <br/>
            Player order: {game.gameState.playerOrder.map(playerId => {
                let isActive = game.activePlayer === playerId;
                return <>
                        <Label basic color={isActive ? 'black' : undefined}>
                            <div style={{
                                height: '1em',
                                width: '1em',
                                borderRadius: '50%',
                                display: 'inline-block',
                                backgroundColor: playerColorToHtml(game.gameState?.playerColor[playerId], userSession)
                            }}/>
                            <LabelDetail>{playerById[playerId].nickname}</LabelDetail>
                        </Label>
                    </>})}<br/>
            Turn: {game.gameState.turnNumber} / {map.getTurnLimit(game.joinedUsers.length)} <br/>
        </Segment>
        <Segment className={"action-holder " + (game.activePlayer === userSession.userInfo?.user.id ? "my-turn" : "other-player-turn") }>
            {actionHolder}
        </Segment>
    </>
}

function UndoSegment({gameId, reload}: {gameId: string, reload: () => Promise<void>}) {
    let [loading, setLoading] = useState<boolean>(false);
    let {setError} = useContext(ErrorContext);

    return <Segment>
        <Button negative loading={loading} onClick={() => {
            setLoading(true);
            return UndoMove({gameId: gameId})
                .then(() => {
                    return reload();
                }).catch(err => {
                    setError(err);
                }).finally(() => {
                    setLoading(false);
                })
        }}>Undo</Button>
    </Segment>
}

function ViewGamePage() {
    let params = useParams();
    let userSession = useContext(UserSessionContext);
    let gameId = params.gameId;

    let [game, setGame] = useState<ViewGameResponse|undefined>(undefined);
    let [gameLogs, setGameLogs] = useState<GetGameLogsResponse|undefined>(undefined);
    let [lastChat, setLastChat] = useState<number>(0);

    const reload: () => Promise<void> = () => {
        userSession.reload();
        if (gameId) {
            return ViewGame({gameId: gameId}).then(res => {
                setGame(res);
            }).then(() => GetGameLogs({gameId: gameId}).then(res => {
                setGameLogs(res);
            }))
        } else {
            setGame(undefined);
            return Promise.resolve();
        }
    };

    useEffect(() => {
        reload();

        let lastMove = 0;
        let pollInterval = setInterval(() => {
            if (gameId) {
                PollGameStatus({gameId: gameId}).then(res => {
                    setLastChat(res.lastChat);
                    if (res.lastMove !== lastMove) {
                        lastMove = res.lastMove;
                        reload();
                    }
                });
            }
        }, 5000);

        return () => clearInterval(pollInterval);
    }, [gameId]);

    if (!game) {
        return <Loader active />
    }

    if (!game.started) {
        let playerCount: number|string;
        if (game.minPlayers === game.maxPlayers) {
            playerCount = game.minPlayers;
        } else {
            playerCount = game.minPlayers + " to " + game.maxPlayers;
        }

        return <>
            <Header as='h1'>Game: {game.name}</Header>
            <Segment>
                <Header as='h2'>Table Info</Header>
                Map: {mapNameToDisplayName(game.mapName)}<br/>
                Player Count: {playerCount}<br/>
                Table Owner: {game.ownerUser.nickname}<br/>
                {game.inviteOnly ? <><span style={{fontStyle: "italic"}}>Invite Only</span><br/></> : null}
            </Segment>
            <Segment>
                <Header as='h2'>Chat</Header>
                <GameChat gameId={game.id} lastChat={lastChat} gameUsers={game.joinedUsers} />
            </Segment>
            <WaitingForPlayersPage game={game} onJoin={() => reload()} onStart={() => reload()} />
        </>
    }

    let canUndo = false;
    if (gameLogs && gameLogs.logs) {
        let maxTimestamp = 0;
        for (let log of gameLogs.logs) {
            if (log.timestamp > maxTimestamp) {
                maxTimestamp = log.timestamp;
                canUndo = (log.reversible && log.userId === userSession.userInfo?.user.id);
            }
        }
    }

    let map = maps[game.mapName];
    let mapInfo = map.getMapInfo();
    return <>
        <Header as='h1'>{game.name} <span style={{fontStyle: "italic"}}>({mapNameToDisplayName(game.mapName)})</span></Header>
        <Segment>
            <Header as='h2'>Chat</Header>
            <GameChat gameId={game.id} lastChat={lastChat} gameUsers={game.joinedUsers} />
        </Segment>
        <PlayerStatus game={game} map={map} onConfirmMove={() => reload()}/>
        {!canUndo ? null : <UndoSegment gameId={game.id} reload={reload} />}
        <ViewMapComponent gameState={game.gameState} activePlayer={game.activePlayer} map={map} />
        <GoodsGrowthTable game={game} map={map} />
        {!mapInfo ? null : <Segment><Header as='h2'>Map Info</Header>{mapInfo}</Segment>}
        <GameLogsComponent game={game} gameLogs={gameLogs} />
        <PartsCountComponent map={map} game={game} />
    </>
}

export default ViewGamePage
