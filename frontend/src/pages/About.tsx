import {Container, Header} from "semantic-ui-react";

function About() {
    return <Container text>
        <Header as='h1'>About</Header>
        <p>This website is a hobby project and there's a lot of UI that might not be intuitive or pretty. I might get to those things eventually. There are no promises made to the availability, uptime, or responsiveness of this site.</p>
        <Header as='h2'>Privacy Policy</Header>
        <p>Your username and email are collected for purposes of authentication and for notifications. Your actions are logged for the purposes of development and debugging, and are generally available to any users of this site. Any data not otherwise accessible by other site users will never be sold or made available to a third-party unless legally obligated.</p>
        <Header as='h2'>Etiquette</Header>
        <p>If you are participating in a game, please be kind to everyone else playing and take your turn in a timely manner. If you are playing with others that you do not know outside of this site, we recommend including expectations around play speed in the table name. Stuff happens and sometimes folks cannot play as fast as intended, and this is also fine. However, please refrain from abandoning incomplete games; that makes everyone sad.</p>
        <Header as='h2'>Contact</Header>
        <p>Questions, comments, requests? Reach out to eot-at-coderealms-dot-io.</p>
    </Container>
}

export default About
