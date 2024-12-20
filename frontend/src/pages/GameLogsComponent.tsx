import {useEffect, useState} from "react";
import {GetGameLogs, GetGameLogsResponse} from "../api/api.ts";
import {Loader, Segment, TextArea} from "semantic-ui-react";

function GameLogsComponent({ gameId, reloadTime }: {gameId: string, reloadTime: number}) {
    let [gameLogs, setGameLogs] = useState<GetGameLogsResponse|undefined>(undefined);
    useEffect(() => {
        GetGameLogs({gameId: gameId}).then(res => {
            setGameLogs(res);
        })
    }, [gameId, reloadTime]);

    return <Segment>
        {gameLogs === undefined ? <Loader active /> :
            <TextArea readOnly disabled style={{width: "100%", height: "30em"}} value={
                gameLogs.logs.map(entry => `[${entry.timestamp}] (${entry.userId}) ${entry.description}`)
                    .join("\n")} />
        }
    </Segment>
}

export default GameLogsComponent
