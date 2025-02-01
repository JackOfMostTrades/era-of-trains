import {Button, Container, Icon, Loader} from "semantic-ui-react";
import {useContext, useEffect, useState} from "react";
import {GameSummary, GetMyGames, GetMyGamesResponse} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {GameSummaryTable} from "../components/GameSummaryTable.tsx";

function MyGames() {
    let [games, setGames] = useState<GetMyGamesResponse|undefined>(undefined);
    let [loading, setLoading] = useState<boolean>(false);
    let {setError} = useContext(ErrorContext);
    let userInfo = useContext(UserSessionContext);

    const reload: () => Promise<void> = () => {
        setLoading(true);
        return GetMyGames({}).then(res => {
            setGames(res);
        }).catch(err => {
            setError(err);
        }).finally(() => {
            setLoading(false);
        })
    }

    useEffect(() => {
        reload();
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
                            && game.joinedUsers.length >= game.minPlayers) {
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
            <div><Button primary floated='right' icon onClick={reload} loading={loading}><Icon name='refresh' /> Refresh</Button><div style={{clear: "both"}} /></div>
            <GameSummaryTable games={waitingForMe} title="Waiting for me" />
            <GameSummaryTable games={activeGames} title="Active Games" />
            <GameSummaryTable games={pendingGames} title="Pending Games" />
            <GameSummaryTable games={finishedGames} title="Finished Games" />
        </Container>
    }
}

export default MyGames
