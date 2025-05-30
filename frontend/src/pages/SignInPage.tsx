import {discordOauthSignin, googleOauthSignin} from "../UserSessionContext.tsx";
import {Button, Container, Header, Icon} from "semantic-ui-react";

function SignInPage() {
    return <Container text>
        <Header as='h1'>Sign-In</Header>
        <p>If you don't have an account yet, sign-in with a provider below and you will be prompted to register.</p>
        <Button primary icon onClick={() => {
            googleOauthSignin('/login/google')
        }}><Icon name="google"/> Sign in with Google</Button>
        <br/>
        <br/>
        <Button primary icon onClick={() => {
            discordOauthSignin('/login/discord')
        }}><Icon name="discord"/> Sign in with Discord</Button>
    </Container>
}

export default SignInPage
