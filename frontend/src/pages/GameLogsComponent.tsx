import {ReactNode, useEffect, useState} from "react";
import {GameLogEntry, GetGameLogs, GetGameLogsResponse, User, ViewGameResponse} from "../api/api.ts";
import {Header, Loader, Segment, TextArea} from "semantic-ui-react";

function renderLogText(playerById: { [playerId: string]: User }, entry: GameLogEntry): string {
    let nick = playerById[entry.userId].nickname;
    let ts = new Date(entry.timestamp*1000).toLocaleString();

    return `[${ts}] (${nick}) ${entry.description}\`)`;
}

function GameLogsComponent({ gameId, game, reloadTime }: {gameId: string, game: ViewGameResponse, reloadTime: number}) {
    let [gameLogs, setGameLogs] = useState<GetGameLogsResponse|undefined>(undefined);
    useEffect(() => {
        GetGameLogs({gameId: gameId}).then(res => {
            setGameLogs(res);
        })
    }, [gameId, reloadTime]);

    let content: ReactNode;
    if (!gameLogs) {
        content = <Loader active />
    } else {
        let playerById: { [playerId: string]: User } = {};
        for (let player of game.joinedUsers) {
            playerById[player.id] = player;
        }

        content = <TextArea readOnly disabled style={{width: "100%", height: "30em"}} value={
            gameLogs.logs?.map(entry => renderLogText(playerById, entry)).join("\n")} />
    }

    return <Segment>
        <Header as='h2'>Logs</Header>
        {content}
    </Segment>
}

export default GameLogsComponent
