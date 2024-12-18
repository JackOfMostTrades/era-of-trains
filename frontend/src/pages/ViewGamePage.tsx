import {Container, Header, Loader} from "semantic-ui-react";
import {useEffect, useState} from "react";
import {useParams} from "react-router";
import {ViewGame, ViewGameResponse} from "../api/api.ts";

function ViewGamePage() {
    let params = useParams();
    let gameId = params.gameId;

    let [game, setGame] = useState<ViewGameResponse|undefined>(undefined);
    useEffect(() => {
        if (gameId) {
            ViewGame({gameId: gameId}).then(res => {
                setGame(res);
            });
        } else {
            setGame(undefined);
        }
    }, [gameId]);

    if (!game) {
        return <Container text style={{marginTop: '7em'}}>
            <Loader active />
        </Container>
    }

    return <Container text style={{marginTop: '7em'}}>
        <Header as='h1'>Game: {game.name}</Header>
    </Container>
}

export default ViewGamePage
