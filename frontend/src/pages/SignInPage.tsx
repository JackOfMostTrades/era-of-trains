import {oauthSignIn} from "../UserSessionContext.tsx";
import {Button, Container, Header} from "semantic-ui-react";

function SignInPage() {
    return <Container text>
        <Header as='h1'>Sign-In</Header>
        <Button primary onClick={() => {
            oauthSignIn()
        }}>Sign in with Google</Button>
    </Container>
}

export default SignInPage
