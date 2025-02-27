import {GameSummary} from "../api/api.ts";
import {Header, Segment, Table, TableBody, TableHeader, TableHeaderCell, TableRow} from "semantic-ui-react";
import GameRow from "./GameRow.tsx";

export function GameSummaryTable({games, title}: {games: GameSummary[], title: string}) {
    if (games.length === 0) {
        return null;
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
                    <TableHeaderCell>Active Player</TableHeaderCell>
                </TableRow>
            </TableHeader>
            <TableBody>{games.map(game => <GameRow key={game.id} game={game} />)}</TableBody>
        </Table>
    </Segment>;
}
