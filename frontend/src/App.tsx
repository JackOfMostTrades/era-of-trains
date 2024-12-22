import {
    Container,
    Dropdown,
    DropdownItem,
    DropdownMenu,
    Icon,
    Label,
    Loader,
    Menu,
    MenuItem,
    MenuMenu,
} from 'semantic-ui-react'
import {BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import UserSessionContext, {logout, UserSessionProvider} from "./UserSessionContext.tsx";
import Home from "./pages/Home.tsx";
import Games from "./pages/Games.tsx";
import {NavLink, useNavigate} from "react-router";
import {useContext} from "react";
import NewGamePage from "./pages/NewGamePage.tsx";
import ViewGamePage from "./pages/ViewGamePage.tsx";
import {ErrorContextProvider} from "./ErrorContext.tsx";
import ErrorDisplay from "./components/ErrorDisplay.tsx";
import About from "./pages/About.tsx";
import DevLogin from "./pages/DevLogin.tsx";
import SignInPage from "./pages/SignInPage.tsx";
import RegisterPage from "./pages/RegisterPage.tsx";
import MyGames from "./pages/MyGames.tsx";
import {createMedia} from '@artsy/fresnel'

const { MediaContextProvider, Media } = createMedia({
    breakpoints: {
        mobile: 0,
        tablet: 768,
        computer: 1024,
    },
})

function UserMenu() {
    let userSessionContext = useContext(UserSessionContext)

    if (userSessionContext.loading) {
        return <MenuItem>
            <Loader active={true} inline={true} size="mini" />
        </MenuItem>
    }

    if (userSessionContext.userInfo) {
        return <>
            <MenuItem>
                <Label color="black">
                    <Icon name='user' /> {userSessionContext.userInfo.user.nickname}
                </Label>
            </MenuItem>
            <MenuItem
                name='Logout'
                onClick={() => logout()}
            />
        </>
    } else {
        return <Menu.Item><NavLink to='/signin'>Sign In</NavLink></Menu.Item>
    }
}

function DesktopMainMenu() {
    let userSessionContext = useContext(UserSessionContext);

    return <Media greaterThan='mobile'>
        <Menu fixed='top' inverted>
            <Container>
                <Menu.Item header>
                    <NavLink to='/'>Era of Trains</NavLink>
                </Menu.Item>
                <Menu.Item header>
                    <NavLink to='/about'>About</NavLink>
                </Menu.Item>
                {userSessionContext.userInfo ? <>
                    <Menu.Item>
                        <NavLink to='/games'>All Games</NavLink>
                    </Menu.Item>
                </> : null}

                <MenuMenu position='right'>
                    {userSessionContext.userInfo ? <>
                        <Menu.Item>
                            {userSessionContext.userInfo.waitingForMeCount ?
                                <Label color='red'>
                                    {userSessionContext.userInfo.waitingForMeCount}
                                </Label> : null}
                            <NavLink to='/mygames'>My Games</NavLink>
                        </Menu.Item>
                    </> : null}
                    <UserMenu />
                </MenuMenu>
            </Container>
        </Menu>
        <Container style={{marginTop: '7em'}} />
    </Media>
}

function MobileMainMenu() {
    let userSessionContext = useContext(UserSessionContext);
    let navigate = useNavigate();

    return <Media at='mobile'>
        <Menu>
            <Menu.Item header>
                Era of Trains
            </Menu.Item>
            <MenuMenu position="right">
                <Dropdown as={Menu.Item} icon="bars">
                    <DropdownMenu>
                        {userSessionContext.userInfo ? <>
                            <DropdownItem disabled>
                                <Label color="black">
                                    <Icon name='user' /> {userSessionContext.userInfo.user.nickname}
                                </Label>
                            </DropdownItem>
                            </> : null }
                        <DropdownItem onClick={() => navigate('/')}>Home</DropdownItem>
                        <DropdownItem onClick={() => navigate('/about')}>About</DropdownItem>
                        {userSessionContext.userInfo ? <>
                        <DropdownItem onClick={() => navigate('/games')}>All Games</DropdownItem>
                            <DropdownItem onClick={() => navigate('/mygames')}>
                                {userSessionContext.userInfo.waitingForMeCount ?
                                    <Label color='red'>
                                        {userSessionContext.userInfo.waitingForMeCount}
                                    </Label> : null}
                                My Games
                            </DropdownItem>

                            <DropdownItem onClick={() => logout()}>Logout</DropdownItem>
                        </> : null}
                        {userSessionContext.loading ?
                            <DropdownItem><Loader active={true} inline={true} size="mini" /></DropdownItem>
                            : null }
                        {!userSessionContext.loading && !userSessionContext.userInfo ?
                            <DropdownItem onClick={() => navigate('/signin')}>Sign In</DropdownItem>
                            : null}
                    </DropdownMenu>
                </Dropdown>
            </MenuMenu>
        </Menu>
        <Container style={{marginTop: '1em'}} />
    </Media>
}

function App() {
    return <MediaContextProvider>
        <ErrorContextProvider>
            <UserSessionProvider>
                <Router>
                    <div>
                        <DesktopMainMenu />
                        <MobileMainMenu />

                        <Container>
                            <ErrorDisplay />
                            <Routes>
                                <Route path="/games/new" element={<NewGamePage />}/>
                                <Route path="/games/:gameId" element={<ViewGamePage />}/>
                                <Route path="/games" element={<Games />}/>
                                <Route path="/mygames" element={<MyGames />}/>
                                <Route path="/about" element={<About />}/>
                                <Route path="/dev/login/:nickname" element={<DevLogin />}/>
                                <Route path="/signin" element={<SignInPage />}/>
                                <Route path="/register" element={<RegisterPage />}/>
                                <Route path="/" element={<Home />}/>
                            </Routes>
                        </Container>
                    </div>
                </Router>
            </UserSessionProvider>
        </ErrorContextProvider>
    </MediaContextProvider>
}

export default App
