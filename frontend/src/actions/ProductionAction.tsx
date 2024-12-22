import {ConfirmMove, Coordinate, ProduceGoodsAction, User, ViewGameResponse} from "../api/api.ts";
import {Button, Grid, GridColumn, GridRow, Header} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {colorToHtml} from "./renderer/HexRenderer.tsx";
import ErrorContext from "../ErrorContext.tsx";

function toCityLabel(n: number): string {
    if (n < 6) {
        return "light city " + (n+1);
    }
    if (n < 12) {
        return "dark city " + (n-5);
    }
    return "new city " + String.fromCharCode(n-12+'A'.charCodeAt(0));
}

function ProductionAction({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [action, setAction] = useState<ProduceGoodsAction>({destinations: []});
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState || !game.gameState.productionCubes) {
        return null;
    }

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail && game.gameState && game.gameState.productionCubes && action.destinations.length < game.gameState.productionCubes.length) {
                let newAction = Object.assign({}, action);
                newAction.destinations = newAction.destinations.slice();
                newAction.destinations.push(e.detail);
                setAction(newAction);
            }
        };

        document.addEventListener('goodsGrowthClickEvent', handler);
        return () => document.removeEventListener('goodsGrowthClickEvent', handler);
    }, [action]);

    let playerById: { [playerId: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.activePlayer) {
        let activePlayer: User|undefined = playerById[game.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to perform the production action...</p>
    } else {

        content = <>
            <p>Click on empty spaces in the goods growth chart where you want to place the cubes, from left to right:</p>
            <Grid>
                <GridRow columns="equal">
                    <GridColumn>
                        <div className="cubeSpot"><div className="cube" style={{background: colorToHtml(game.gameState.productionCubes[0])}}/></div>
                        {action.destinations.length >= 1 ? <p>Placing on {toCityLabel(action.destinations[0].x)} in spot {action.destinations[0].y+1}.</p> : null}
                    </GridColumn>
                    <GridColumn>
                        <div className="cubeSpot"><div className="cube" style={{background: colorToHtml(game.gameState.productionCubes[1])}}/></div>
                        {action.destinations.length >= 2 ? <p>Placing on {toCityLabel(action.destinations[1].x)} in spot {action.destinations[1].y+1}.</p> : null}
                    </GridColumn>
                </GridRow>
            </Grid>
            <br/>
            <Button primary disabled={action.destinations.length < game.gameState.productionCubes.length} loading={loading} onClick={() => {
                setLoading(true);
                ConfirmMove({
                    gameId: game.id,
                    actionName: "produce_goods",
                    produceGoodsAction: action,
                }).then(() => {
                    return onDone();
                }).catch(err => {
                    setError(err);
                }).finally(() => {
                    setLoading(false);
                });
            }}>Confirm</Button>
            <Button negative onClick={() => setAction({destinations: []})}>Restart</Button>
        </>;
    }

    return <>
        <Header as='h2'>Goods Growth</Header>
        {content}
    </>
}

export default ProductionAction
