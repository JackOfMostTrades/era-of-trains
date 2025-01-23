import {Button, Header, Loader, Table, TableBody, TableHeader, TableHeaderCell, TableRow} from "semantic-ui-react";
import {useEffect, useState} from "react";
import {Link} from "react-router";
import {ListGames, ListGamesResponse} from "../api/api.ts";
import GameRow from "../components/GameRow.tsx";

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
        if (games.games) {
            table = <Table celled>
                <TableHeader>
                    <TableRow>
                        <TableHeaderCell>Name</TableHeaderCell>
                        <TableHeaderCell>Number of Players</TableHeaderCell>
                        <TableHeaderCell>Map</TableHeaderCell>
                        <TableHeaderCell>Status</TableHeaderCell>
                    </TableRow>
                </TableHeader>
                <TableBody>{games.games.map(game => <GameRow key={game.id} game={game} />)}</TableBody>
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
