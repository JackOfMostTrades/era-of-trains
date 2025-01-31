import {Button, Header, Loader} from "semantic-ui-react";
import {useEffect, useState} from "react";
import {Link} from "react-router";
import {GameSummary, ListGames, ListGamesResponse} from "../api/api.ts";
import {GameSummaryTable} from "../components/GameSummaryTable.tsx";

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
            let availableGames: GameSummary[] = [];
            let activeGames: GameSummary[] = [];
            let finishedGames: GameSummary[] = [];

            for (let game of games.games) {
                if (!game.started && game.joinedUsers.length < game.maxPlayers) {
                    availableGames.push(game);
                } else if (game.finished) {
                    finishedGames.push(game);
                } else {
                    activeGames.push(game);
                }
            }

            table = <>
                <GameSummaryTable games={availableGames} title="Waiting For Players" />
                <GameSummaryTable games={activeGames} title="Active Games" />
                <GameSummaryTable games={finishedGames} title="Finished Games" />
            </>
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
