import {
    Container,

    Icon, Label, Loader,
    Menu, MenuItem, MenuMenu,
} from 'semantic-ui-react'
import {BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import {logout, oauthSignIn, UserSessionProvider} from "./UserSessionContext.tsx";
import Home from "./pages/Home.tsx";
import Games from "./pages/Games.tsx";
import {NavLink} from "react-router";
import {useContext} from "react";
import UserSessionContext from "./UserSessionContext.tsx";
import NewGamePage from "./pages/NewGamePage.tsx";
import ViewGamePage from "./pages/ViewGamePage.tsx";

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
                    <Icon name='user' /> {userSessionContext.userInfo.User}
                </Label>
            </MenuItem>
            <MenuItem
                name='Logout'
                onClick={() => logout()}
            />
        </>
    } else {
        return <MenuItem
            name='Sign In'
            onClick={() => oauthSignIn()}
        />
    }
}

function MainMenu() {
    let userSessionContext = useContext(UserSessionContext);

    return <Menu fixed='top' inverted>
        <Container>
            <Menu.Item header>
                <NavLink to='/'>Era of Trains</NavLink>
            </Menu.Item>
            {userSessionContext.userInfo ? <>
                <Menu.Item>
                    <NavLink to='/games'>Games</NavLink>
                </Menu.Item>
            </> : null}

            <MenuMenu position='right'>
                <UserMenu />
            </MenuMenu>
        </Container>
    </Menu>
}

function App() {
    return <UserSessionProvider>
        <Router>
            <div>
                <MainMenu />

                <Routes>
                    <Route path="/games/new" element={<NewGamePage />}/>
                    <Route path="/games/:gameId" element={<ViewGamePage />}/>
                    <Route path="/games" element={<Games />}/>
                    <Route path="/" element={<Home />}/>
                </Routes>
            </div>
        </Router>
    </UserSessionProvider>
}

export default App
