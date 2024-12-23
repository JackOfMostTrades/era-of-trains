import {Container, Header, List, ListItem} from "semantic-ui-react";

function Home() {
    return <Container text>
        <Header as='h1'>Era of Trains</Header>
        <p>Currently in private beta, so there's not a lot of exposition or tutorial here at present. If you're in the private beta and need help or hit a bug, ping me on Discord.</p>
        <Header as='h2'>Known Issues</Header>
        <List>
            <ListItem>Component limits are not enforced (town markers, number of tiles, etc.)</ListItem>
            <ListItem>Bankrupt players do not get eliminated (they just go into negative income)</ListItem>
        </List>
    </Container>
}

export default Home
