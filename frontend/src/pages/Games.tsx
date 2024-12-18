import {
    Button,
    Container,
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

interface GameSummary {
    id: string;
    name: string;
}
interface ListGamesResponse {
    games?: GameSummary[];
}

function Games() {
    let [games, setGames] = useState<ListGamesResponse|undefined>(undefined);
    useEffect(() => {
        (async () => {
            let res = await fetch('/api/listGames', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({})
            })
            if (!res.ok) {
                throw new Error("got non-ok response");
            }
            setGames(await res.json());
        })();
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
                    <TableCell>Click me</TableCell>
                </TableRow>)}
            </TableBody>
        </Table>;
    }

    return <Container text style={{marginTop: '7em'}}>
        <Header as='h1'>Games</Header>
        <Button primary>Start New Game</Button>
        {table}
    </Container>
}

export default Games
