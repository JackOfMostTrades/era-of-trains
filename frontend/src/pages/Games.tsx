import {
    Button,
    Header,
    Loader,
    Table,
    TableBody,
    TableCell,
    TableHeader,
    TableHeaderCell,
    TableRow
} from "semantic-ui-react";
import {ReactNode, useEffect, useState} from "react";
import {Link} from "react-router";
import {ListGames, ListGamesResponse} from "../api/api.ts";

function Games() {
    let [games, setGames] = useState<ListGamesResponse|undefined>(undefined);
    useEffect(() => {
        ListGames({}).then(res => {
            setGames(res);
        }).catch(err => {
            console.error(err);
        });
    }, []);

    let table;
    if (!games) {
        table = <Loader active={true} />
    } else {
        let rows: ReactNode[] = [];
        if (games.games) {
            for (let game of games.games) {
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
                    <TableCell>{game.mapName}</TableCell>
                    <TableCell>{status}</TableCell>
                </TableRow>)
            }

            table = <Table celled>
                <TableHeader>
                    <TableRow>
                        <TableHeaderCell>Name</TableHeaderCell>
                        <TableHeaderCell>Number of Players</TableHeaderCell>
                        <TableHeaderCell>Map</TableHeaderCell>
                        <TableHeaderCell>Status</TableHeaderCell>
                    </TableRow>
                </TableHeader>
                <TableBody>{rows}</TableBody>
            </Table>;
        }
    }

    return <>
        <Header as='h1'>Games</Header>
        <Link to="/games/new">
            <Button primary>Start New Game</Button>
        </Link>
        {table}
    </>
}

export default Games
