import {Button, Grid, Header, Label, LabelDetail, List, ListItem, Loader, Segment} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import {useParams} from "react-router";
import {
    GamePhase,
    JoinGame,
    LeaveGame,
    PlayerColor,
    PollGameStatus,
    StartGame,
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

function WaitingForPlayersPage({game, onJoin}: {game: ViewGameResponse, onJoin: () => Promise<void>}) {
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
    for (let i = game.joinedUsers.length; i < game.numPlayers; i++) {
        listItems.push(<ListItem></ListItem>);
    }

    return <>
        <p>Waiting for players...</p>
        <List>{listItems}</List>
        {joined ? <>
                <Button negative loading={loading} onClick={() => {
                    setLoading(true)
                    LeaveGame({gameId: game.id}).then(() => {
                        return onJoin();
                    }).finally(() => {
                        setLoading(false);
                    })
                }}>Leave Game</Button>
            </> : <>
                <Button primary loading={loading} onClick={() => {
                    setLoading(true)
                    JoinGame({gameId: game.id}).then(() => {
                        return onJoin();
                    }).finally(() => {
                        setLoading(false);
                    })
                }}>Join Game</Button>
            </>}
    </>
}

function WaitingForStartPage({game, onStart}: {game: ViewGameResponse, onStart: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [loading, setLoading] = useState<boolean>(false);

    let listItems: ReactNode[] = [];
    for (let i = 0; i < game.joinedUsers.length; i++) {
        let player = game.joinedUsers[i];
        listItems.push(<ListItem>{player.nickname}</ListItem>);
    }

    return <>
        <p>Waiting for owner to start game...</p>
        <List>{listItems}</List>
        <Button primary loading={loading} disabled={userSession.userInfo?.user.id !== game.ownerUser.id} onClick={() => {
            setLoading(true)
            StartGame({gameId: game.id}).then(() => {
                return onStart();
            }).finally(() => {
                setLoading(false);
            })
        }}>Start Game</Button>
    </>
}

function PlayerColorAndName({nickname, color}: {nickname: string, color: PlayerColor|undefined}) {
    return <><div style={{
        height: '1em',
        width: '1em',
        borderRadius: '50%',
        display: 'inline-block',
        backgroundColor: playerColorToHtml(color)
    }}/> {nickname}</>
}

function PlayerStatus({game, map, onConfirmMove}: { game: ViewGameResponse, map: GameMap, onConfirmMove: () => Promise<void> }) {
    let userSessionContext = useContext(UserSessionContext);

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
        playerColumns.push(<Grid.Column key={playerId}><Segment>
                Player: <PlayerColorAndName nickname={nickname} color={game.gameState.playerColor[playerId]} /><br/>
                Cash: ${game.gameState.playerCash[playerId]}<br/>
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
                                backgroundColor: playerColorToHtml(game.gameState?.playerColor[playerId])
                            }}/>
                            <LabelDetail>{playerById[playerId].nickname}</LabelDetail>
                        </Label>
                    </>})}<br/>
            Turn: {game.gameState.turnNumber} / {map.getTurnLimit(game.numPlayers)} <br/>
        </Segment>
        <Segment className={"action-holder " + (game.activePlayer === userSessionContext.userInfo?.user.id ? "my-turn" : "other-player-turn") }>
            {actionHolder}
        </Segment>
    </>
}

function ViewGamePage() {
    let params = useParams();
    let userSession = useContext(UserSessionContext);
    let gameId = params.gameId;

    let [game, setGame] = useState<ViewGameResponse|undefined>(undefined);
    let [reloadTime, setReloadTime] = useState<number>(0);
    let [lastChat, setLastChat] = useState<number>(0);

    const reload: () => Promise<void> = () => {
        userSession.reload();
        if (gameId) {
            return ViewGame({gameId: gameId}).then(res => {
                setGame(res);
                setReloadTime(Date.now());
            });
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
        let content: ReactNode
        if (game.joinedUsers.length < game.numPlayers) {
            content = <WaitingForPlayersPage game={game} onJoin={() => reload()} />
        } else {
            content = <WaitingForStartPage game={game} onStart={() => reload()}/>
        }
        return <>
            <Header as='h1'>Game: {game.name}</Header>
            <Segment>
                <Header as='h2'>Table Info</Header>
                Map: {mapNameToDisplayName(game.mapName)}<br/>
                Player Count: {game.numPlayers}<br/>
                Table Owner: {game.ownerUser.nickname}<br/>
            </Segment>
            <Segment>
                <Header as='h2'>Chat</Header>
                <GameChat gameId={game.id} lastChat={lastChat} gameUsers={game.joinedUsers} />
            </Segment>
            {content}
        </>
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
        <ViewMapComponent game={game} map={map} />
        <GoodsGrowthTable game={game} map={map} />
        {!mapInfo ? null : <Segment><Header as='h2'>Map Info</Header>{mapInfo}</Segment>}
        <GameLogsComponent gameId={game.id} game={game} reloadTime={reloadTime} />
    </>
}

export default ViewGamePage
