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
import {useEffect, useState} from "react";
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
        table = <Table celled>
            <TableHeader>
                <TableRow>
                    <TableHeaderCell>Id</TableHeaderCell>
                    <TableHeaderCell>Name</TableHeaderCell>
                    <TableHeaderCell>View</TableHeaderCell>
                </TableRow>
            </TableHeader>
            <TableBody>
                {games.games?.map(game => <TableRow>
                    <TableCell>{game.id}</TableCell>
                    <TableCell>{game.name}</TableCell>
                    <TableCell><Link to={`/games/${game.id}`}>Click me</Link></TableCell>
                </TableRow>)}
            </TableBody>
        </Table>;
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
