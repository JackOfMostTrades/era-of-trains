import {ConfirmMove, User, ViewGameResponse} from "../api/api.ts";
import {
    Button,
    Dropdown,
    DropdownItemProps,
    Header,
    List,
    ListItem,
    Modal,
    ModalActions,
    ModalContent,
    ModalDescription,
    ModalHeader
} from "semantic-ui-react";
import {ReactNode, useContext, useState} from "react";
import UserSessionContext from "../UserSessionContext.tsx";
import ErrorContext from "../ErrorContext.tsx";

function ConfirmBidWithTopModal({open, onConfirm, onCancel}: {open: boolean, onConfirm: () => void, onCancel: () => void}) {
    return (
        <Modal open={open}>
            <ModalHeader>Skip using turn order pass?</ModalHeader>
            <ModalContent>
                <ModalDescription>
                    <Header>You have turn order pass</Header>
                    <p>You have turn order pass. Do you really want to skip using it?</p>
                </ModalDescription>
            </ModalContent>
            <ModalActions>
                <Button primary onClick={onConfirm}>Yes, continue with bid</Button>
                <Button negative onClick={onCancel}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

function ConfirmPassWithTopModal({open, onConfirm, onCancel}: {open: boolean, onConfirm: () => void, onCancel: () => void}) {
    return (
        <Modal open={open}>
            <ModalHeader>Skip using turn order pass?</ModalHeader>
            <ModalContent>
                <ModalDescription>
                    <Header>You have turn order pass</Header>
                    <p>You have turn order pass. Do you really want to pass instead of using it?</p>
                </ModalDescription>
            </ModalContent>
            <ModalActions>
                <Button primary onClick={onConfirm}>Yes, pass and give up turn order pass</Button>
                <Button negative onClick={onCancel}>Cancel</Button>
            </ModalActions>
        </Modal>
    )
}

function AuctionAction({game, onDone}: {game: ViewGameResponse, onDone: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let {setError} = useContext(ErrorContext);
    let [amount, setAmount] = useState<number>(0);
    let [loading, setLoading] = useState<boolean>(false);
    let [showBidWithTopModal, setShowBidWithTopModal] = useState<boolean>(false);
    let [showPassWithTopModal, setShowPassWithTopModal] = useState<boolean>(false);

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
        if (bid < 0) {
            let position = game.gameState.playerOrder.length + bid + 1;
            currentBids.push(<ListItem>{player.nickname}: <span style={{fontStyle: "italic"}}>passed: {position}</span></ListItem>)
        } else {
            currentBids.push(<ListItem>{player.nickname}: {bid}</ListItem>)
        }
    }

    let content: ReactNode;

    if (userSession.userInfo?.user.id !== game.activePlayer) {
        let activePlayer: User|undefined = playerById[game.activePlayer];
        content = <p>Waiting for {activePlayer?.nickname} to bid...</p>
    } else {
        let currentCash = game.gameState.playerCash[game.activePlayer];
        let currentMaxBid = 0;
        for (let playerId of game.gameState.playerOrder) {
            let bid = 0;
            if (game.gameState.auctionState) {
                bid = game.gameState.auctionState[playerId] || 0;
            }
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
            }).catch(err => {
                setError(err);
            }).finally(() => {
                setLoading(false);
            });
        }

        let hasTurnOrderPass = game.gameState.playerActions[game.activePlayer] === 'turn_order_pass';
        let turnOrderPassButton: ReactNode;
        if (hasTurnOrderPass) {
            turnOrderPassButton = <><Button secondary loading={loading} onClick={() => doBid(0)}>Turn Order Pass</Button></>
        }

        content = <>
            <p>Choose your bid: </p>
            <Dropdown selection
                      value={amount}
                      onChange={(_, {value}) => setAmount(value as number)}
                      options={options}/><br/>
            <Button primary loading={loading} onClick={() => {
                if (hasTurnOrderPass) {
                    setShowBidWithTopModal(true);
                } else {
                    doBid(amount)
                }
            }}>Bid</Button>
            {turnOrderPassButton}
            <Button negative loading={loading} onClick={() => {
                if (hasTurnOrderPass) {
                    setShowPassWithTopModal(true);
                } else {
                    doBid(-1)
                }
            }}>Pass</Button>
            <ConfirmBidWithTopModal open={showBidWithTopModal} onConfirm={() => {
                setShowBidWithTopModal(false);
                doBid(amount);
            }} onCancel={() => {
                setShowBidWithTopModal(false);
            }} />
            <ConfirmPassWithTopModal open={showPassWithTopModal} onConfirm={() => {
                setShowPassWithTopModal(false);
                doBid(-1);
            }} onCancel={() => {
                setShowPassWithTopModal(false);
            }} />
        </>;
    }

    return <>
        <Header as='h2'>Auction</Header>
        <List>{currentBids}</List>
        {content}
    </>
}

export default AuctionAction