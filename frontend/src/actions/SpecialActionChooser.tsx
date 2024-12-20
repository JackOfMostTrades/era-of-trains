import {ConfirmMove, SpecialAction, User, ViewGameResponse} from "../api/api.ts";
import {Button, Dropdown, DropdownItemProps, Header} from "semantic-ui-react";
import {ReactNode, useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";

function SpecialActionChooser({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [action, setAction] = useState<SpecialAction|undefined>(undefined);
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    let options: DropdownItemProps[] = [
        {
            key: "first_move",
            value: "first_move",
            text: "First Move",
        },
        {
            key: "first_build",
            value: "first_build",
            text: "First Build",
        },
        {
            key: "engineer",
            value: "engineer",
            text: "Engineer",
        },
        {
            key: "loco",
            value: "loco",
            text: "Locomotive",
        },
        {
            key: "urbanization",
            value: "urbanization",
            text: "Urbanization",
        },
        {
            key: "production",
            value: "production",
            text: "Production",
        },
        {
            key: "turn_order_pass",
            value: "turn_order_pass",
            text: "Turn-Order Pass"
        },
    ]
    for (let playerId of Object.keys(game.gameState.playerActions)) {
        let action = game.gameState.playerActions[playerId];
        if (action) {
            for (let i = 0; i < options.length; i++) {
                if (options[i].value === action) {
                    options.splice(i, 1);
                    break;
                }
            }
        }
    }

    let playerById: { [playerId: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.gameState.activePlayer) {
        let activePlayer: User|undefined = playerById[game.gameState.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to choose an action...</p>
    } else {
        content = <>
            <p>Choose your action: </p>
            <Dropdown selection
                      value={action}
                      onChange={(_, {value}) => setAction(value as SpecialAction)}
                      options={options}/><br/>
            <Button primary loading={loading} onClick={() => {
                setLoading(true);
                ConfirmMove({
                    gameId: game.id,
                    actionName: "choose_action",
                    chooseAction: {
                        action: action,
                    }
                }).then(() => {
                    return onDone();
                }).catch(err => {
                    setError(err);
                }).finally(() => {
                    setLoading(false);
                });
            }}>Select Action</Button><br/>
        </>;
    }

    return <>
        <Header as='h2'>Choosing Special Actions</Header>
        {content}
    </>
}

export default SpecialActionChooser
