import {ConfirmMove, GameState, SpecialAction, User, ViewGameResponse} from "../api/api.ts";
import {
    Button,
    Dropdown,
    DropdownItemProps,
    Header,
    Modal, ModalActions,
    ModalContent,
    ModalDescription,
    ModalHeader
} from "semantic-ui-react";
import {ReactNode, useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";
import {specialActionToDisplayName} from "../util.ts";

function isNewCityAvailable(gameState: GameState|undefined): boolean {
    if (!gameState || !gameState.urbanizations) {
        return true;
    }
    return gameState.urbanizations.length !== 8;
}

function ConfirmUrbModal({open, onConfirm, onCancel}: {open: boolean, onConfirm: () => void, onCancel: () => void}) {
    return (
        <Modal open={open}>
            <ModalHeader>Skip Actions?</ModalHeader>
            <ModalContent>
                <ModalDescription>
                    <Header>There are no new cities left</Header>
                    <p>You are selecting the urbanization action, but there are no new cities left that can be placed. Are you sure you want to pick this action?</p>
                </ModalDescription>
            </ModalContent>
            <ModalActions>
                <Button primary onClick={onConfirm}>Yes, pick urbanization</Button>
                <Button negative onClick={onCancel}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

function SpecialActionChooser({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [action, setAction] = useState<SpecialAction|undefined>(undefined);
    let [loading, setLoading] = useState<boolean>(false);
    let [showConfirmModal, setShowConfirmModal] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    const commitAction = () => {
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
                if (action === 'urbanization' && !isNewCityAvailable(game.gameState)) {
                    setShowConfirmModal(true);
                    return;
                }
                commitAction();
            }}>Select Action</Button><br/>
        </>;
    }

    return <>
        <Header as='h2'>Choosing Special Actions</Header>
        <ConfirmUrbModal
            open={showConfirmModal}
            onConfirm={() => {
                setShowConfirmModal(false);
                commitAction();
            }}
            onCancel={() => setShowConfirmModal(false)} />
        {content}
    </>
}

export default SpecialActionChooser
