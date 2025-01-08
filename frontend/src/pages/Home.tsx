import {Container, Header, List, ListItem} from "semantic-ui-react";
import {Link} from "react-router";

function Home() {
    return <Container text>
        <Header as='h1'>Era of Trains</Header>
        <p>Era of Trains is a site for playing a particular series of train games that are set in the early age of the steam locomotive. This site is a hobby project developed and maintained in free time, and is free for anyone to use. Games on this site are intended for play in an asynchronous, turn-based manner. This site does not have a rules explanation or tutorial mode, but anyone who has played at least once before should find taking their turn easy enough to follow.</p>
        <p>To get started, you can head over to the <Link to='/signin'>sign-in page</Link> and you'll be prompted to create an account. Once you've created your account, head over to the <Link to='/games'>All Games</Link> page. You can also view any finished game or game in progress to get a sense of the interface. From the games page, feel free to join a game that's waiting for players, or feel free to start a new table. The volume of users on this site is low, so organizing players for a game on other social channels is recommended if you're going to start a new table.</p>
        <p>The functionality of this site should be considered "beta" as a number of minor issues have been discovered (and fixed) during initial tests. Please proceed with modest expectations.</p>
        <Header as='h2'>Announcements</Header>
        <Header as='h3'>Jan 7, 2025</Header>
        <List>
            <ListItem>Added support for signing in with Discord.</ListItem>
            <ListItem>Added support for receiving a Discord webhook when it's your turn.</ListItem>
            <ListItem>Various minor bugfixes and UI improvements.</ListItem>
            <ListItem>Text changes to indicate movement to public beta.</ListItem>
        </List>
        <Header as='h3'>Dec 30, 2024</Header>
        <List>
            <ListItem>Initial release of the Scotland map is available.</ListItem>
            <ListItem>Maps have rivers now!</ListItem>
            <ListItem>Users with pending turns will now receive a daily summary email as a reminder.</ListItem>
            <ListItem>Cubes glide over the map during move goods action. So fancy!</ListItem>
        </List>
        <Header as='h3'>Dec 26, 2024</Header>
        <List>
            <ListItem>Initial release of the Germany map is available.</ListItem>
            <ListItem>Bankrupt players are now eliminated.</ListItem>
            <ListItem>Users can now set player color preferences that get applied when starting a new game.</ListItem>
            <ListItem>Number of steps now displayed during move goods actions.</ListItem>
            <ListItem>Component limits (track tiles and town markers) are now enforced during builds.</ListItem>
            <ListItem>Map-specific info is now displayed on game pages.</ListItem>
        </List>
        <Header as='h3'>Dec 24, 2024</Header>
        <List>
            <ListItem>Initial release of the Southern U.S. map is available.</ListItem>
            <ListItem>You can now redirect incomplete tracks during builds.</ListItem>
            <ListItem>Fixed a bug allowing deliveries to repeat cities.</ListItem>
            <ListItem>Many other miscellaneous bug fixes and improvements.</ListItem>
        </List>
        <Header as='h3'>Dec 20, 2024</Header>
        <p>Initial private beta release. Let the games begin!</p>
    </Container>
}

export default Home
