import {Button, Container, Header, List, ListItem, Loader, Segment, SegmentGroup} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import {useParams} from "react-router";
import {GamePhase, JoinGame, LeaveGame, StartGame, User, ViewGame, ViewGameResponse} from "../api/api.ts";
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

function PlayerStatus({ game, onConfirmMove }: {game: ViewGameResponse, onConfirmMove: () => Promise<void>}) {
    if (!game.gameState) {
        return null;
    }

    let playerById: { [id: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }
    let playerOrder: string[] = [];
    for (let playerId of game.gameState.playerOrder) {
        let player = playerById[playerId];
        playerOrder.push(player.nickname);
    }

    let playerColumns: ReactNode[] = [];
    for (let player of game.joinedUsers) {
        let playerColorHtml = playerColorToHtml(game.gameState.playerColor[player.id]);
        playerColumns.push(<Segment key={player.id}>
                Player: <div style={{height: '1em', width: '1em', borderRadius: '50%', display: 'inline-block', backgroundColor: playerColorHtml}} /> {player.nickname}<br/>
                Cash: ${game.gameState.playerCash[player.id]}<br/>
                Shares: {game.gameState.playerShares[player.id]}<br/>
                Income: {game.gameState.playerIncome[player.id]}<br/>
                Loco: {game.gameState.playerLoco[player.id]}<br/>
                Special Action: {game.gameState.playerActions[player.id]}<br/>
            </Segment>);
    }

    let actionHolder: ReactNode;
    if (game.finished) {
        actionHolder = <FinalScore game={game} />
    } else {
        switch (game.gameState.gamePhase) {
            case GamePhase.SHARES:
                actionHolder = <ChooseShares game={game} onDone={onConfirmMove}/>
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
            <Header as='h2'>Player Info</Header>
            <SegmentGroup horizontal>
                {playerColumns}
            </SegmentGroup>
        </Segment>
        <Container>
            <Segment>
                <Header as='h2'>Game Status</Header>
                Player order: {playerOrder.join(", ")}<br/>
                Active player: {playerById[game.gameState.activePlayer].nickname}<br/>
                Game Phase: {game.gameState.gamePhase}<br/>
                Turn: {game.gameState.turnNumber}<br/>
            </Segment>
            <Segment>
                {actionHolder}
            </Segment>
        </Container>
    </>
}

function ViewGamePage() {
    let params = useParams();
    let gameId = params.gameId;

    let [game, setGame] = useState<ViewGameResponse|undefined>(undefined);
    let [reloadTime, setReloadTime] = useState<number>(0);

    const reload: () => Promise<void> = () => {
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
    }, [gameId]);

    if (!game) {
        return <Loader active />
    }

    let content: ReactNode
    if (!game.started) {
        if (game.joinedUsers.length < game.numPlayers) {
            content = <WaitingForPlayersPage game={game} onJoin={() => reload()} />
        } else {
            content = <WaitingForStartPage game={game} onStart={() => reload()}/>
        }
    } else {
        content = <>
            <PlayerStatus game={game} onConfirmMove={() => reload()}/>
            <ViewMapComponent game={game} />
            <GoodsGrowthTable game={game} />
            <GameLogsComponent gameId={game.id} game={game} reloadTime={reloadTime} />
        </>
    }

    return <>
        <Header as='h1'>Game: {game.name}</Header>
        <Segment>
            <Header as='h2'>Table Info</Header>
            Map: {game.mapName}<br/>
            Player Count: {game.numPlayers}<br/>
            Table Owner: {game.ownerUser.nickname}<br/>
        </Segment>
        {content}
    </>
}

export default ViewGamePage
