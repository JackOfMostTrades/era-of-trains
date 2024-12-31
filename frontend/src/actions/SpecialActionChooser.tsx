import {ConfirmMove, SpecialAction, User, ViewGameResponse} from "../api/api.ts";
import {Button, Dropdown, DropdownItemProps, Header} from "semantic-ui-react";
import {ReactNode, useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {specialActionToDisplayName} from "../util.ts";

function SpecialActionChooser({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [action, setAction] = useState<SpecialAction|undefined>(undefined);
    let [loading, setLoading] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    let availableSpecialActions: SpecialAction[] = ['first_move', 'first_build', 'engineer', 'loco', 'urbanization', 'production', 'turn_order_pass'];
    for (let playerId of Object.keys(game.gameState.playerActions)) {
        let action = game.gameState.playerActions[playerId];
        if (action) {
            for (let i = 0; i < availableSpecialActions.length; i++) {
                if (availableSpecialActions[i] === action) {
                    availableSpecialActions.splice(i, 1);
                    break;
                }
            }
        }
    }
    let options: DropdownItemProps[] = availableSpecialActions
        .map(specialAction => ({
            key: specialAction,
            value: specialAction,
            text: specialActionToDisplayName(specialAction)
        }))

    let playerById: { [playerId: string]: User } = {};
    for (let player of game.joinedUsers) {
        playerById[player.id] = player;
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.activePlayer) {
        let activePlayer: User|undefined = playerById[game.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to choose an action...</p>
    } else {
        content = <>
            <p>Choose your action: </p>
            <Dropdown selection
                      value={action}
                      onChange={(_, {value}) => setAction(value as SpecialAction)}
                      options={options}/><br/>
            <Button primary loading={loading} disabled={!action} onClick={() => {
                setLoading(true);
                ConfirmMove({
                    gameId: game.id,
                    actionName: "choose_action",
                    chooseAction: {
                        action: action as SpecialAction,
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
