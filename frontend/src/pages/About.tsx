import {Container, Header} from "semantic-ui-react";

function About() {
    return <Container text>
        <Header as='h1'>About</Header>
        <p>This website is a hobby project and there's a lot of UI that might not be intuitive or pretty. I might get to those things eventually. There are no promises made to the availability, uptime, or responsiveness of this site.</p>
        <Header as='h2'>Privacy Policy</Header>
        <p>Your username and email are collected for purposes of authentication and for notifications. Your actions are logged for the purposes of development and debugging, and are generally available to any users of this site. Any data not otherwise accessible by other site users will never be sold or made available to a third-party unless legally obligated.</p>
        <Header as='h2'>Etiquette</Header>
        <p>If you are participating in a game, please be kind to everyone else playing. Keep the discussion friendly; no harassment will be tolerated. Please try and take your turns in a timely manner; if you create a table we recommend including expectations around play speed in the table name. Stuff happens and sometimes folks cannot play as fast as intended. Stuff happens to everyone, just let folks know in the game chat if you'll be delayed. Also, refrain from abandoning incomplete games; that makes everyone sad.</p>
        <Header as='h2'>Contact</Header>
        <p>Questions, comments, requests? <a target="_blank" href="https://discord.gg/f9rtBrhW2j">Discord</a> is where most chat happens, but you can also reach out with email to eot-at-coderealms-dot-io.</p>
    </Container>
}

export default About
