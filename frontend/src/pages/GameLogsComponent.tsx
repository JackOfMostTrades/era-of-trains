import {ReactNode} from "react";
import {GameLogEntry, GetGameLogsResponse, User, ViewGameResponse} from "../api/api.ts";
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
    TableRow,
} from "semantic-ui-react";

function LogRow({playerById, entry}: {playerById: { [playerId: string]: User }, entry: GameLogEntry}) {
    let nick = playerById[entry.userId].nickname;
    let ts = new Date(entry.timestamp*1000).toLocaleString();

    return <TableRow>
        <TableCell>{ts}</TableCell>
        <TableCell>{nick}</TableCell>
        <TableCell><div style={{whiteSpace: "pre-line"}}>{entry.description}</div></TableCell>
    </TableRow>
}

function GameLogsComponent({ game, gameLogs }: {game: ViewGameResponse, gameLogs: GetGameLogsResponse|undefined}) {
    let content: ReactNode;
    if (!gameLogs) {
        content = <Loader active />
    } else {
        let playerById: { [playerId: string]: User } = {};
        for (let player of game.joinedUsers) {
            playerById[player.id] = player;
        }

        let entries: ReactNode[] = [];
        if (gameLogs.logs) {
            for (let idx = gameLogs.logs.length-1; idx >= 0; idx--) {
                entries.push(<LogRow key={idx} playerById={playerById} entry={gameLogs.logs[idx]} />);
            }
        }

        content = <Container>
                <div style={{height: "30em", overflowY: "scroll"}}>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHeaderCell>When</TableHeaderCell>
                                <TableHeaderCell>Who</TableHeaderCell>
                                <TableHeaderCell>What</TableHeaderCell>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {entries}
                        </TableBody>
                    </Table>
                </div>
            </Container>
    }

    return <Segment>
        <Header as='h2'>Logs</Header>
        {content}
    </Segment>
}

export default GameLogsComponent
