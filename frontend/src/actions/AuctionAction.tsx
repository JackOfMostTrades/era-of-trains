import {ConfirmMove, User, ViewGameResponse} from "../api/api.ts";
import {Button, Dropdown, DropdownItemProps, Header, List, ListItem} from "semantic-ui-react";
import {ReactNode, useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";

function AuctionAction({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [amount, setAmount] = useState<number>(0);
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    let playerById: { [playerId: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let currentBids: ReactNode[] = [];
    for (let playerId of game.gameState.playerOrder) {
        let player = playerById[playerId];
        let bid = 0;
        if (game.gameState.auctionState) {
            bid = game.gameState.auctionState[playerId] || 0;
        }
        currentBids.push(<ListItem>{player.nickname}: {bid}</ListItem>)
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.gameState.activePlayer) {
        let activePlayer: User|undefined = playerById[game.gameState.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to bid...</p>
    } else {
        let currentCash = game.gameState.playerCash[game.gameState.activePlayer];
        let currentMaxBid = 0;
        for (let playerId of game.gameState.playerOrder) {
            let bid = game.gameState.auctionState[playerId] || 0;
            currentMaxBid = Math.max(currentMaxBid, bid);
        }

        let options: DropdownItemProps[] = [];
        for (let i = currentMaxBid+1; i <= currentCash; i++) {
            options.push({
                key: i,
                text: i,
                value: i,
            })
        }

        const doBid = (val: number) => {
            setLoading(true);
            ConfirmMove({
                gameId: game.id,
                actionName: "bid",
                bidAction: {
                    amount: val,
                }
            }).then(() => {
                return onDone();
            }).finally(() => {
                setLoading(false);
            });
        }

        let turnOrderPassButton: ReactNode;
        if (game.gameState.playerActions[game.gameState.activePlayer] === 'turn_order_pass') {
            turnOrderPassButton = <><Button secondary loading={loading} onClick={() => doBid(0)} /><br/></>
        }

        content = <>
            <p>Choose your bid: </p>
            <Dropdown selection
                      value={amount}
                      onChange={(_, {value}) => setAmount(value as number)}
                      options={options}/><br/>
            <Button primary loading={loading} onClick={() => doBid(amount)}>Bid</Button><br/>
            {turnOrderPassButton}
            <Button negative loading={loading} onClick={() => doBid(-1)}>Pass</Button>
        </>;
    }

    return <>
        <Header as='h2'>Auction</Header>
        <List>{currentBids}</List>
        {content}
    </>
}

export default AuctionAction