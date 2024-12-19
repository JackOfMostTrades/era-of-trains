import {ConfirmMove, User, ViewGameResponse} from "../api/api.ts";
import {Button, Dropdown, DropdownItemProps} from "semantic-ui-react";
import {useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";

function ChooseShares({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [amount, setAmount] = useState<number>(0);
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    if (userSession.userInfo?.user.id !== game.gameState.activePlayer) {
        let activePlayer: User|undefined;
        for (let player of game.joinedUsers) {
            if (player.id === game.gameState.activePlayer) {
                activePlayer = player;
                break;
            }
        }

        return <p>Waiting for {activePlayer?.nickname} to choose number of shares...</p>
    }

    let currentShares = game.gameState.playerShares[game.gameState.activePlayer];
    let options: DropdownItemProps[] = [];
    for (let i = 0; i <= 15-currentShares; i++) {
        options.push({
            key: i,
            text: i,
            value: i,
        })
    }

    return <>
        <p>Choose how many shares: </p>
        <Dropdown selection
            value={amount}
            onChange={(_, {value}) => setAmount(value as number)}
            options={options} />
        <Button primary loading={loading} onClick={() => {
            setLoading(true);
            ConfirmMove({
                gameId: game.id,
                actionName: "shares",
                sharesAction: {
                    amount: amount,
                }
            }).then(() => {
                return onDone();
            }).finally(() => {
                setLoading(false);
            })
        }}>Take shares</Button>
    </>;
}

export default ChooseShares