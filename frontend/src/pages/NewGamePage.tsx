import {Button, Checkbox, Dropdown, Form, FormField, Header, Input} from "semantic-ui-react";
import {useNavigate} from "react-router";
import {CreateGame, CreateGameRequest} from "../api/api.ts";
import {useState} from "react";
import {mapNameToDisplayName} from "../util.ts";

function NewGamePage() {
    const navigate = useNavigate();
    let [req, setReq] = useState<CreateGameRequest>({
        name: "",
        minPlayers: 2,
        maxPlayers: 6,
        mapName: "rust_belt",
        inviteOnly: false,
    });
    let [loading, setLoading] = useState<boolean>(false);

    return <>
        <Header as='h1'>New Game</Header>
        <Form>
            <FormField>
                <label>Name</label>
                <Input placeholder='Some short description...' value={req.name} onChange={(_, { value }) => {
                    let newReq = Object.assign({}, req);
                    newReq.name = value;
                    setReq(newReq);
                }} />
            </FormField>
            <FormField>
                <label>Minimum number of players</label>
                <Dropdown
                    selection
                    value={req.minPlayers}
                    onChange={(_, { value }) => {
                        let newReq = Object.assign({}, req);
                        newReq.minPlayers = value as number;
                        newReq.maxPlayers = Math.max(newReq.maxPlayers, newReq.minPlayers);
                        setReq(newReq);
                    }}
                    options={[
                        {
                            key: "2",
                            text: "2",
                            value: 2,
                        },
                        {
                            key: "3",
                            text: "3",
                            value: 3,
                        },
                        {
                            key: "4",
                            text: "4",
                            value: 4,
                        },
                        {
                            key: "5",
                            text: "5",
                            value: 5,
                        },
                        {
                            key: "6",
                            text: "6",
                            value: 6,
                        }
                    ]}
                />
            </FormField>
            <FormField>
                <label>Maximum number of players</label>
                <Dropdown
                    selection
                    value={req.maxPlayers}
                    onChange={(_, { value }) => {
                        let newReq = Object.assign({}, req);
                        newReq.maxPlayers = value as number;
                        newReq.minPlayers = Math.min(newReq.minPlayers, newReq.maxPlayers);
                        setReq(newReq);
                    }}
                    options={[
                        {
                            key: "2",
                            text: "2",
                            value: 2,
                        },
                        {
                            key: "3",
                            text: "3",
                            value: 3,
                        },
                        {
                            key: "4",
                            text: "4",
                            value: 4,
                        },
                        {
                            key: "5",
                            text: "5",
                            value: 5,
                        },
                        {
                            key: "6",
                            text: "6",
                            value: 6,
                        }
                    ]}
                />
            </FormField>
            <FormField>
                <label>Map</label>
                <Dropdown
                    selection
                    value={req.mapName}
                    onChange={(_, { value }) => {
                        let newReq = Object.assign({}, req);
                        newReq.mapName = value as string;
                        setReq(newReq);
                    }}
                    options={["rust_belt", "southern_us", "germany", "scotland", "australia"].map(mapName => ({
                        key: mapName,
                        value: mapName,
                        text: mapNameToDisplayName(mapName)
                    }))}
                />
            </FormField>
            <FormField>
                <label>Invite-Only</label>
                <p>Games marked as invite-only will not be listed on the "All Games" page (but it will show up on your "My Games" page). You will need to send a link to the game to whomever you want to have join.</p>
                <Checkbox toggle checked={req.inviteOnly} onChange={(_, val) => {
                    let newReq = Object.assign({}, req);
                    newReq.inviteOnly = !!val.checked;
                    setReq(newReq);
                }} />
            </FormField>
            <Button primary loading={loading} type='submit' onClick={() => {
                setLoading(true);
                CreateGame(req).then(res => {
                    navigate(`/games/${res.id}`);
                }).catch(err => {
                    console.error(err);
                }).finally(() => {
                    setLoading(false);
                })
            }}>Create</Button>
        </Form>
    </>
}

export default NewGamePage
