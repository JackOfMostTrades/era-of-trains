import {Header} from "semantic-ui-react";

function About() {
    return <>
        <Header as='h1'>About</Header>
        <p>This website is a hobby project. There's a lot of UI that's not intuitive or not pretty. I might get to those things eventually. There are no promises made to the availability, uptime, or responsiveness of this site.</p>
        <Header as='h2'>Privacy Policy</Header>
        <p>Your username and email are collected for purposes of authentication and for notifications. Your actions are logged for the purposes of development and debugging, and are generally available to any users of this site. Any data not otherwise accessible by other site users will never be sold or made available to a third-party unless legally obligated.</p>
    </>
}

export default About
