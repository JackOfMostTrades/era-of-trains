import {ConfirmMove, Coordinate, ProduceGoodsAction, User, ViewGameResponse} from "../api/api.ts";
import {Button, Grid, GridColumn, GridRow, Header, List} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import {colorToHtml} from "./renderer/HexRenderer.tsx";

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
    let [action, setAction] = useState<ProduceGoodsAction>({destination: []});
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState || !game.gameState.productionCubes) {
        return null;
    }

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate>) => {
            if (e.detail && action.destination.length < game.gameState.productionCubes.length) {
                let newAction = Object.assign({}, action);
                newAction.destination = newAction.destination.slice();
                newAction.destination.push(e.detail);
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

    if (userSession.userInfo?.user.id !== game.gameState.activePlayer) {
        let activePlayer: User|undefined = playerById[game.gameState.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to perform the production action...</p>
    } else {

        content = <>
            <p>Click on empty spaces in the goods growth chart where you want to place the cubes, from left to right:</p>
            <Grid>
                <GridRow columns="equal">
                    <GridColumn>
                        <div className="cubeSpot"><div className="cube" color={colorToHtml(game.gameState.productionCubes[0])}/></div>
                        {action.destination.length >= 1 ? <p>Placing on {toCityLabel(action.destination[0].x)} in spot {action.destination[0].y+1}.</p> : null}
                    </GridColumn>
                    <GridColumn>
                        <div className="cubeSpot"><div className="cube" color={colorToHtml(game.gameState.productionCubes[1])}/></div>
                        {action.destination.length >= 2 ? <p>Placing on {toCityLabel(action.destination[1].x)} in spot {action.destination[1].y+1}.</p> : null}
                    </GridColumn>
                </GridRow>
            </Grid>
            <Button primary disabled={action.destination.length < game.gameState.productionCubes.length} loading={loading} onClick={() => {
                setLoading(true);
                ConfirmMove({
                    gameId: game.id,
                    actionName: "produce_goods",
                    produceGoodsAction: action,
                }).then(() => {
                    return onDone();
                }).finally(() => {
                    setLoading(false);
                });
            }}>Confirm</Button><br/>
            <Button negative onClick={() => setAction({destination: []})}>Restart</Button>
        </>;
    }

    return <>
        <Header as='h2'>Auction</Header>
        <List>{currentBids}</List>
        {content}
    </>
}

export default ProductionAction
