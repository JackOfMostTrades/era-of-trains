import {discordOauthSignin, googleOauthSignin} from "../UserSessionContext.tsx";
import {Button, Container, Header, Icon} from "semantic-ui-react";

function SignInPage() {
    return <Container text>
        <Header as='h1'>Sign-In</Header>
        <p>If you don't have an account yet, you need to initially sign-in with Google to register.</p>
        <Button primary icon onClick={() => {
            googleOauthSignin()
        }}><Icon name="google"/> Sign in with Google</Button>
        <br/>
        <br/>
        <p>Link your profile with your Discord account to be able to sign-in with Discord.</p>
        <Button primary icon onClick={() => {
            discordOauthSignin('/login/discord')
        }}><Icon name="discord"/> Sign in with Discord</Button>
    </Container>
}

export default SignInPage
