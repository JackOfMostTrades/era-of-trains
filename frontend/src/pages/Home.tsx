import {Container, Header} from "semantic-ui-react";

function Home() {
    return <Container text>
        <Header as='h1'>Era of Trains</Header>
        <p>Era of Trains is a site for playing a particular series of train games that are set in the early age of the steam locomotive. This site is a hobby project developed and maintained in free time, and is free for anyone to use.</p>
        <p>This site is now in "maintenance mode" in favor of further development over at <a href="https://www.choochoo.games">choochoo.games</a>. Whether you're new here or a veteran player, we recommended you head over there as it's where new maps and features will be implemented.</p>
        <p>Questions or comments? Find us on <a target="_blank" href="https://discord.gg/f9rtBrhW2j">Discord</a>! This is a good place to find other players, get notifications if it's your turn, or report issues.</p>
    </Container>
}

export default Home
