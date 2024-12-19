import {Button, Grid, GridColumn, GridRow, Header, List, ListItem, Loader} from "semantic-ui-react";
import {ReactNode, useContext, useEffect, useState} from "react";
import {useParams} from "react-router";
import {Color, Coordinate, JoinGame, LeaveGame, StartGame, User, ViewGame, ViewGameResponse} from "../api/api.ts";
import UserSessionContext from "../UserSessionContext.tsx";
import maps, {BasicMap, HexType} from "../map.ts"

function WaitingForPlayersPage({game, onJoin}: {game: ViewGameResponse, onJoin: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [loading, setLoading] = useState<boolean>(false);

    let joined = false;
    for (let player of game?.joinedUsers) {
        if (player.id === userSession.userInfo?.user.id) {
            joined = true;
            break;
        }
    }

    let listItems: ReactNode[] = [];
    for (let i = 0; i < game.joinedUsers.length; i++) {
        let player = game.joinedUsers[i];
        listItems.push(<ListItem>{player.nickname}</ListItem>);
    }
    for (let i = game.joinedUsers.length; i < game.numPlayers; i++) {
        listItems.push(<ListItem></ListItem>);
    }

    return <>
        <p>Waiting for players...</p>
        <List>{listItems}</List>
        {joined ? <>
                <Button negative loading={loading} onClick={() => {
                    setLoading(true)
                    LeaveGame({gameId: game.id}).then(() => {
                        return onJoin();
                    }).finally(() => {
                        setLoading(false);
                    })
                }}>Leave Game</Button>
            </> : <>
                <Button primary loading={loading} onClick={() => {
                    setLoading(true)
                    JoinGame({gameId: game.id}).then(() => {
                        return onJoin();
                    }).finally(() => {
                        setLoading(false);
                    })
                }}>Join Game</Button>
            </>}
    </>
}

function WaitingForStartPage({game, onStart}: {game: ViewGameResponse, onStart: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [loading, setLoading] = useState<boolean>(false);

    let listItems: ReactNode[] = [];
    for (let i = 0; i < game.joinedUsers.length; i++) {
        let player = game.joinedUsers[i];
        listItems.push(<ListItem>{player.nickname}</ListItem>);
    }

    return <>
        <p>Waiting for owner to start game...</p>
        <List>{listItems}</List>
        <Button primary loading={loading} disabled={userSession.userInfo?.user.id !== game.ownerUser.id} onClick={() => {
            setLoading(true)
            StartGame({gameId: game.id}).then(() => {
                return onStart();
            }).finally(() => {
                setLoading(false);
            })
        }}>Start Game</Button>
    </>
}

function colorToHtml(color: Color): string {
    switch (color) {
        case Color.NONE: return '#ffffff';
        case Color.BLACK: return "#444444";
        case Color.BLUE: return '#00a2d3';
        case Color.RED: return '#d41e2d';
        case Color.PURPLE: return '#8d5a95';
        case Color.YELLOW: return '#f5ac11';
    }
}

interface CityState {
    label: string
    color: Color
    darkCity: boolean
}

function coordinateToKey(coordinate: Coordinate): string {
    return coordinate.x + "," + coordinate.y;
}

function getCityState(game: ViewGameResponse, map: BasicMap, coordinate: Coordinate): CityState|undefined {
    if (game.gameState && game.gameState.urbanizations) {
        for (let urb of game.gameState.urbanizations) {
            if (urb.hex.x === coordinate.x && urb.hex.y === coordinate.y) {
                let label: string;
                let color: Color;
                let darkCity: boolean;
                switch (urb.city) {
                    case 0:
                        label = 'A';
                        color = Color.RED;
                        darkCity = false;
                        break;
                    case 1:
                        label = 'B';
                        color = Color.BLUE;
                        darkCity = false;
                        break;
                    case 2:
                        label = 'C';
                        color = Color.BLACK;
                        darkCity = false;
                        break;
                    case 3:
                        label = 'D';
                        color = Color.BLACK;
                        darkCity = false;
                        break;
                    case 4:
                        label = 'E';
                        color = Color.YELLOW;
                        darkCity = true;
                        break;
                    case 5:
                        label = 'F';
                        color = Color.PURPLE;
                        darkCity = true;
                        break;
                    case 6:
                        label = 'G';
                        color = Color.BLACK;
                        darkCity = true;
                        break;
                    case 7:
                        label = 'H';
                        color = Color.BLACK;
                        darkCity = true;
                        break;
                }
                return {
                    label: label,
                    color: color,
                    darkCity: darkCity
                };
            }
        }
    }

    for (let city of map.cities) {
        if (city.coordinate.x === coordinate.x && city.coordinate.y === coordinate.y) {
            let label = city.goodsGrowth.map(n => (n%6)+1).join(',');
            let color = city.color;
            let darkCity = false;
            for (let goodsGrowth of city.goodsGrowth) {
                if (goodsGrowth >= 6) {
                    darkCity = true;
                    break;
                }
            }
            return {
                label: label,
                color: color,
                darkCity: darkCity
            };
        }
    }

    return undefined;
}

function ViewMapPage({game, onUpdate}: {game: ViewGameResponse, onUpdate: () => Promise<void>}) {
    let userSession = useContext(UserSessionContext);
    let [lastClick, setLastClick] = useState<Coordinate>({x: 0, y: 0});
    let activePlayer: User|undefined;
    for (let player of game.joinedUsers) {
        if (player.id === game.gameState.activePlayer) {
            activePlayer = player;
            break;
        }
    }
    let myTurn = false;
    if (activePlayer && userSession.userInfo?.user.id === activePlayer.id) {
        myTurn = true;
    }

    let map = maps[game.mapName];
    let paths: ReactNode[] = []
    for (let y = 0; y < map.height; y++) {
        for (let x = 0; x < map.width; x++) {
            let hex = map.hexes[y][x];
            if (hex == HexType.OFFBOARD) {
                continue;
            }

            let cityState: CityState|undefined = getCityState(game, map, {x: x, y: y});

            let color: string;
            switch (hex) {
                case HexType.WATER:
                    color = '#579ba8';
                    break;
                case HexType.RIVER:
                    color = '#537845';
                    break;
                case HexType.PLAINS:
                    color = '#99c37b';
                    break;
                case HexType.MOUNTAIN:
                    color = '#867565';
                    break;
                case HexType.TOWN:
                    color = '#99c37b';
                    break;
                case HexType.CITY:
                    color = '#ffffff';
                    if (cityState && cityState.darkCity) {
                        color = '#222222';
                    }
                    break;
            }
            let xpos = x*17.321;
            if ((y % 2) === 1) {
                xpos += 8.661;
            }
            let ypos = y*5;

            const onClick = () => setLastClick({x: x, y: y});

            let points = `${xpos},${ypos+5} ${xpos+2.887},${ypos} ${xpos+8.661},${ypos} ${xpos+11.547},${ypos+5} ${xpos+8.661},${ypos+10} ${xpos+2.887},${ypos+10}`
            paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

            if (hex == HexType.TOWN) {
                paths.push(<circle cx={xpos+5.7735} cy={ypos+5} r={2.5} fill='#FFFFFF' onClick={onClick} />);
            } else if (hex == HexType.CITY) {
                if (cityState) {
                    let cityColor = colorToHtml(cityState.color);
                    let strokeColor: string = '#222222';
                    if (cityState.darkCity) {
                        strokeColor = '#ffffff';
                    }
                    let points = `${xpos + 0.8},${ypos + 5} ${xpos + 3.225},${ypos + 0.8} ${xpos + 8.075},${ypos + 0.8} ${xpos + 10.747},${ypos + 5} ${xpos + 8.075},${ypos + 9.2} ${xpos + 3.225},${ypos + 9.2}`
                    paths.push(<polygon stroke={strokeColor} strokeWidth={0.2} fill={cityColor} points={points} onClick={onClick}/>);
                    paths.push(<circle cx={xpos + 5.7735} cy={ypos + 5} r={2.5} fill='#FFFFFF' onClick={onClick}/>);
                    paths.push(<text fontSize={2.5} x={xpos + 5.7735} y={ypos+5.3} dominant-baseline="middle" text-anchor="middle">{cityState.label}</text>);
                }
            }
        }
    }

    if (game.gameState) {
        if (game.gameState.links) {
            for (let link of game.gameState.links) {
                // FIXME: Render links
            }
        }

        if (game.gameState.cubes) {
            let coordByKey: { [key: string]: Coordinate } = {};
            let cubesByKey: { [key: string]: Color[] } = {};
            for (let cube of game.gameState.cubes) {
                let key = `${cube.hex.x},${cube.hex.y}`
                coordByKey[key] = cube.hex;
                if (!cubesByKey[key]) {
                    cubesByKey[key] = [];
                }
                cubesByKey[key].push(cube.color);
            }

            for (let key of Object.keys(coordByKey)) {
                let hex = coordByKey[key];
                let cubes = cubesByKey[key];

                let xpos = hex.x*17.321;
                if ((hex.y % 2) === 1) {
                    xpos += 8.661;
                }
                let ypos = hex.y*5;

                // Center the xpos
                xpos += (11.547-cubes.length*2.5+0.5)/2;

                for (let i = 0; i < cubes.length; i++) {
                    let cube = cubes[i];
                    let points = `${xpos+i*2.5},${ypos+0.5} ${xpos+2+i*2.5},${ypos+0.5} ${xpos+2+i*2.5},${ypos+2.5} ${xpos+i*2.5},${ypos+2.5}`
                    paths.push(<polygon stroke='#222222' strokeWidth={0.25} fill={colorToHtml(cube)} points={points} filter="url(#cube-shadow)"/>);
                }
            }
        }
    }

    let mapRender = <svg
        xmlns="http://www.w3.org/2000/svg"
        width={(map.width * 17.321 + 8.661) * 6}
        height={map.height * 60}
        viewBox={`0 0 ${map.width * 17.321 + 8.661} ${map.height * 10}`}>
        <defs>
            <filter id="cube-shadow" width="2.5" height="2.5">
                <feOffset in="SourceAlpha" dx="0.5" dy="0.5"/>
                <feGaussianBlur stdDeviation="0.25"/>
                <feBlend in="SourceGraphic" in2="blurOut"/>
            </filter>
        </defs>
        {paths}
    </svg>;

    return <>
        <p>Active player: {activePlayer?.nickname}</p>
        <p>Last click: {JSON.stringify(lastClick)}</p>
        {mapRender}
    </>
}

function ViewGamePage() {
    let params = useParams();
    let gameId = params.gameId;

    let [game, setGame] = useState<ViewGameResponse|undefined>(undefined);

    const reload: () => Promise<void> = () => {
        if (gameId) {
            return ViewGame({gameId: gameId}).then(res => {
                setGame(res);
            });
        } else {
            setGame(undefined);
            return Promise.resolve();
        }
    };

    useEffect(() => {
        reload();
    }, [gameId]);

    if (!game) {
        return <Loader active />
    }

    let content: ReactNode
    if (!game.started) {
        if (game.joinedUsers.length < game.numPlayers) {
            content = <WaitingForPlayersPage game={game} onJoin={() => reload()} />
        } else {
            content = <WaitingForStartPage game={game} onStart={() => reload()}/>
        }
    } else {
        content = <ViewMapPage game={game} onUpdate={() => reload()}/>
    }

    return <>
        <Header as='h1'>Game: {game.name}</Header>
        <Grid columns="equal">
            <GridRow>
                <GridColumn>Map</GridColumn>
                <GridColumn>{game.mapName}</GridColumn>
            </GridRow>
            <GridRow>
                <GridColumn>Player Count</GridColumn>
                <GridColumn>{game.numPlayers}</GridColumn>
            </GridRow>
            <GridRow>
                <GridColumn>Table Owner</GridColumn>
                <GridColumn>{game.ownerUser.nickname}</GridColumn>
            </GridRow>
        </Grid>
        {content}
    </>
}

export default ViewGamePage
