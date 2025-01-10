import {useContext, useEffect, useState} from "react";
import {GameChatMessage, GetGameChat, SendGameChat, User} from "../api/api.ts";
import {Form, Input} from "semantic-ui-react";
import UserSessionContext from "../UserSessionContext.tsx";
import "./GameChat.css";

interface Props {
    gameId: string;
    lastChat: number;
    gameUsers: User[];
}

function GameChat(props: Props) {
    let {userInfo} = useContext(UserSessionContext);
    let [lastRefresh, setLastRefresh] = useState<{gameId: string, after: number}>({gameId: "", after: 0});
    let [messages, setMessages] = useState<GameChatMessage[]>([]);
    let [toSend, setToSend] = useState<string>("");
    let [sending, setSending] = useState<boolean>(false);

    const refreshNow = () => {
        let gameId = props.gameId;
        let after = lastRefresh.after;
        if (props.gameId !== lastRefresh.gameId) {
            after = 0;
        }

        GetGameChat({gameId: gameId, after: after}).then(res => {
            let newMessages = messages.slice();
            if (res.messages) {
                newMessages = newMessages.concat(res.messages);
            }
            setMessages(newMessages);
            if (newMessages.length > 0) {
                setLastRefresh({gameId: gameId, after: newMessages[newMessages.length - 1].timestamp});
            }
        });
    };

    useEffect(() => {
        refreshNow();
    }, [props.gameId, props.lastChat]);

    let playerIdToNickMap: { [playerId: string]: string } = {};
    for (let user of props.gameUsers) {
        playerIdToNickMap[user.id] = user.nickname;
    }

    return <div className="chat">
        <div className="chatlog" ref={(element) => {
            if (element) {
                element.scrollTop = element.scrollHeight
            }
        }}>
            {messages.map(message => {
                let ts = new Date(message.timestamp * 1000).toLocaleString();
                return <div key={message.timestamp} className="chatline">
                    <span className="timestamp">{ts}</span>
                    <span className="nickname">{playerIdToNickMap[message.userId] || message.userId}</span>{' '}
                    <span className="message">{message.message}</span>
                </div>
            })}
        </div>
        {userInfo?.user?.id && playerIdToNickMap[userInfo.user.id] ? <>
            <div className="chatbar">
                <Form onSubmit={() => {
                    setSending(true);
                    SendGameChat({gameId: props.gameId, message: toSend}).then(() => {
                        setSending(false);
                        setToSend("");
                        return refreshNow();
                    }).catch(_ => {
                        setSending(false);
                    });
                }}>
                    <Input fluid disabled={sending} value={toSend} placeholder="Type something and press enter..."
                           onChange={(_, data) => setToSend(data.value)}/>
                </Form>
            </div>
        </> : null}
    </div>
}

export default GameChat;
