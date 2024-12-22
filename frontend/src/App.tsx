import {Container, Icon, Label, Loader, Menu, MenuItem, MenuMenu,} from 'semantic-ui-react'
import {BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import UserSessionContext, {logout, UserSessionProvider} from "./UserSessionContext.tsx";
import Home from "./pages/Home.tsx";
import Games from "./pages/Games.tsx";
import {NavLink} from "react-router";
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

function MainMenu() {
    let userSessionContext = useContext(UserSessionContext);

    return <Menu fixed='top' inverted>
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
}

function App() {
    return <ErrorContextProvider>
        <UserSessionProvider>
            <Router>
                <div>
                    <MainMenu />

                    <Container style={{marginTop: '7em'}}>
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
}

export default App
