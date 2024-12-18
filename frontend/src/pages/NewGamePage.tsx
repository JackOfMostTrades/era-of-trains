import {Button, Container, Dropdown, Form, FormField, Header, Input} from "semantic-ui-react";
import {useNavigate} from "react-router";
import {CreateGame, CreateGameRequest} from "../api/api.ts";
import {useState} from "react";

function NewGamePage() {
    const navigate = useNavigate();
    let [req, setReq] = useState<CreateGameRequest>({
        name: "",
        numPlayers: 2,
        mapName: "rust_belt",
    });
    let [loading, setLoading] = useState<boolean>(false);

    return <Container text style={{marginTop: '7em'}}>
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
                <label>Number of players</label>
                <Dropdown
                    selection
                    value={req.numPlayers}
                    onChange={(_, { value }) => {
                        let newReq = Object.assign({}, req);
                        newReq.numPlayers = value as number;
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
                    options={[
                        {
                            key: "rust_belt",
                            text: "Rust Belt",
                            value: "rust_belt",
                        },
                    ]}
                />
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
    </Container>
}

export default NewGamePage
