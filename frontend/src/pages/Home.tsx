import {Container, Header} from "semantic-ui-react";
import {Link} from "react-router";

function Home() {
    return <Container text>
        <Header as='h1'>Era of Trains</Header>
        <p>Era of Trains is a site for playing a particular series of train games that are set in the early age of the steam locomotive. This site is a hobby project developed and maintained in free time, and is free for anyone to use. Games on this site are intended for play in an asynchronous, turn-based manner. This site does not have a rules explanation or tutorial mode, but anyone who has played at least once before should find taking their turn easy enough to follow.</p>
        <p>To get started, you can head over to the <Link to='/signin'>sign-in page</Link> and you'll be prompted to create an account. Once you've created your account, head over to the <Link to='/games'>All Games</Link> page. You can also view any finished game or game in progress to get a sense of the interface. From the games page, feel free to join a game that's waiting for players or start a new table.</p>
        <p>The volume of users on this site is low, so organizing players for a game on other social channels is recommended. We've started a dedicated Discord server, which you can join with <a target="_blank" href="https://discord.gg/f9rtBrhW2j">this link</a>! This is a good place to find other players, get notifications if it's your turn, or report issues.</p>
        <p>Announcements, discussion, and feature requests all happen on Discord, so it's highly recommended that you join us there!</p>
    </Container>
}

export default Home
