import {
    Container,
    Header,
    Loader,
    Segment,
    Table,
    TableBody,
    TableCell,
    TableHeader,
    TableHeaderCell,
    TableRow
} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import {Link} from "react-router";
import {GameSummary, GetMyGames, GetMyGamesResponse} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {mapNameToDisplayName} from "../util.ts";

function GameSummaryTable({games, title}: {games: GameSummary[], title: string}) {
    if (games.length === 0) {
        return null;
    }

    let rows: ReactNode[] = [];
    for (let game of games) {
        let status: string;
        if (!game.started) {
            if (game.joinedUsers.length < game.numPlayers) {
                status = 'Waiting for players';
            } else {
                status = 'Waiting to start';
            }
        } else if (game.finished) {
            status = 'Finished';
        } else {
            status = 'In progress';
        }

        rows.push(<TableRow>
            <TableCell><Link to={`/games/${game.id}`}>{game.name}</Link></TableCell>
            <TableCell>{game.numPlayers}</TableCell>
            <TableCell>{mapNameToDisplayName(game.mapName)}</TableCell>
            <TableCell>{status}</TableCell>
        </TableRow>)
    }

    return <Segment>
        <Header as='h2'>{title}</Header>
        <Table celled>
            <TableHeader>
                <TableRow>
                    <TableHeaderCell>Name</TableHeaderCell>
                    <TableHeaderCell>Number of Players</TableHeaderCell>
                    <TableHeaderCell>Map</TableHeaderCell>
                    <TableHeaderCell>Status</TableHeaderCell>
                </TableRow>
            </TableHeader>
            <TableBody>{rows}</TableBody>
        </Table>
    </Segment>;
}

function MyGames() {
    let [games, setGames] = useState<GetMyGamesResponse|undefined>(undefined);
    let {setError} = useContext(ErrorContext);
    let userInfo = useContext(UserSessionContext);
    useEffect(() => {
        GetMyGames({}).then(res => {
            setGames(res);
        }).catch(err => {
            setError(err);
        });
    }, []);

    if (!games) {
        return <Loader active={true} />
    } else {
        let waitingForMe: GameSummary[] = [];
        let activeGames: GameSummary[] = [];
        let pendingGames: GameSummary[] = [];
        let finishedGames: GameSummary[] = [];

        if (games.games) {
            for (let game of games.games) {
                if (!game.started) {
                    if (game.ownerUser.id === userInfo.userInfo?.user.id
                            && game.joinedUsers.length === game.numPlayers) {
                        waitingForMe.push(game);
                    } else {
                        pendingGames.push(game);
                    }
                } else {
                    if (game.finished) {
                        finishedGames.push(game);
                    } else {
                        if (game.activePlayer === userInfo.userInfo?.user.id) {
                            waitingForMe.push(game);
                        } else {
                            activeGames.push(game);
                        }
                    }
                }
            }
        }

        return <Container>
            <GameSummaryTable games={waitingForMe} title="Waiting for me" />
            <GameSummaryTable games={activeGames} title="Active Games" />
            <GameSummaryTable games={pendingGames} title="Pending Games" />
            <GameSummaryTable games={finishedGames} title="Finished Games" />
        </Container>
    }
}

export default MyGames
