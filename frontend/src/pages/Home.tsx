import {Container, Header, List, ListItem} from "semantic-ui-react";

function Home() {
    return <Container text>
        <Header as='h1'>Era of Trains</Header>
        <p>Currently only available by invitation.</p>
        <Header as='h2'>Known Issues</Header>
        <List>
            <ListItem>Track replacement of incomplete track is currently not possible</ListItem>
            <ListItem>Component limits are not enforced (town markers, number of tiles, etc.)</ListItem>
            <ListItem>Bankrupt players do not get eliminated (they just go into negative income)</ListItem>
        </List>
    </Container>
}

export default Home
