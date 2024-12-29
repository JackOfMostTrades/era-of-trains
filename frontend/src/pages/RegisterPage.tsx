import {Button, Container, Form, FormField, Header, Input} from "semantic-ui-react";
import {useContext, useState} from "react";
import {Register} from "../api/api.ts";
import ErrorContext from "../ErrorContext.tsx";
import {useNavigate} from "react-router";

function SignInPage() {
    let {setError} = useContext(ErrorContext);
    let navigate = useNavigate();
    let [nickname, setNickname] = useState<string>("");

    return <Container text>
        <Header as='h1'>Register</Header>
        <p>Looks like you don't have an account yet. Fill this in to get started!</p>
        <Form>
            <FormField>
                <label>Nickname</label>
                <p>This is how you'll show up in games. Letters and numbers only.</p>
                <Input placeholder='Nickname' value={nickname} onChange={(_, { value }) => {
                    setNickname(value);
                }} />
            </FormField>
            <Button primary onClick={() => {
                let params = new URLSearchParams(window.location.hash.substring(1));
                let accessToken = params.get("access_token");
                if (accessToken) {
                    Register({
                        accessToken: accessToken,
                        nickname: nickname,
                    }).then(() => {
                        return navigate('/profile');
                    }).then(() => {
                        window.location.reload();
                    }).catch(err => {
                        setError(err);
                    })
                }
            }}>Create Account</Button>
        </Form>

    </Container>
}

export default SignInPage
