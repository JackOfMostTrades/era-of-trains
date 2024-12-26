import {Color, Coordinate, Direction, PlayerColor} from "../../api/api.ts";
import {CityProperties, HexType} from "../../maps";
import {ReactNode} from "react";

export function colorToHtml(color: Color): string {
    switch (color) {
        case Color.NONE: return '#ffffff';
        case Color.BLACK: return "#444444";
        case Color.BLUE: return '#00a2d3';
        case Color.RED: return '#d41e2d';
        case Color.PURPLE: return '#8d5a95';
        case Color.YELLOW: return '#f5ac11';
        case Color.WHITE: return '#e9dcc9';
    }
}

export function playerColorToHtml(color: PlayerColor|undefined): string {
    if (color === undefined) {
        return '#222222';
    }
    switch (color) {
        case PlayerColor.BLUE: return '#00839d';
        case PlayerColor.GREEN: return "#a7ed36";
        case PlayerColor.YELLOW: return '#f3db70';
        case PlayerColor.PINK: return '#bf88a9';
        case PlayerColor.GRAY: return '#96a496';
        case PlayerColor.ORANGE: return '#db7e2a';
    }
}

export function urbCityProperties(n: number): CityProperties {
    let label: string;
    let color: Color;
    let darkCity: boolean;
    switch (n) {
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
        default:
            throw new Error("unhandled urbanization city: " + n);
    }
    return {
        label: label,
        color: color,
        darkCity: darkCity
    };
}

function hexEdgeOffset(direction: Direction): {dx: number, dy: number} {
    switch (direction) {
        case Direction.NORTH:
            return {dx: 5.7735, dy: 0};
        case Direction.NORTH_EAST:
            return {dx: 10.104, dy: 2.5};
        case Direction.SOUTH_EAST:
            return {dx: 10.104, dy: 7.5};
        case Direction.SOUTH:
            return {dx: 5.7735, dy: 10};
        case Direction.SOUTH_WEST:
            return {dx: 1.4435, dy: 7.5};
        case Direction.NORTH_WEST:
            return {dx: 1.4435, dy: 2.5};
    }
}

export class HexRenderer {
    private width: number;
    private height: number;
    private paths: ReactNode[];
    private emitOnClick: boolean;
    private filterId: number;
    private static nextFilterId: number = 0;

    constructor(emitOnClock: boolean) {
        this.width = -1;
        this.height = -1;
        this.paths = [];
        this.emitOnClick = emitOnClock;
        this.filterId = HexRenderer.nextFilterId;
        HexRenderer.nextFilterId += 1;
    }

    public renderCityHex(hex: Coordinate, cityProperties: CityProperties) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        let onClick: undefined | (() => void);
        if (this.emitOnClick) {
            onClick = () => {
                let event = new CustomEvent('mapClickEvent', {
                    detail: {
                        x: hex.x,
                        y: hex.y,
                    }
                });
                document.dispatchEvent(event);
            }
        }

        let color = '#ffffff';
        if (cityProperties.darkCity) {
            color = '#222222';
        }

        let points = `${xpos},${ypos+5} ${xpos+2.887},${ypos} ${xpos+8.661},${ypos} ${xpos+11.547},${ypos+5} ${xpos+8.661},${ypos+10} ${xpos+2.887},${ypos+10}`
        this.paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

        let cityColor = colorToHtml(cityProperties.color);
        let strokeColor: string = '#222222';
        if (cityProperties.darkCity) {
            strokeColor = '#ffffff';
        }
        points = `${xpos + 0.8},${ypos + 5} ${xpos + 3.225},${ypos + 0.8} ${xpos + 8.075},${ypos + 0.8} ${xpos + 10.747},${ypos + 5} ${xpos + 8.075},${ypos + 9.2} ${xpos + 3.225},${ypos + 9.2}`
        this.paths.push(<polygon stroke={strokeColor} strokeWidth={0.2} fill={cityColor} points={points} onClick={onClick}/>);
        this.paths.push(<circle cx={xpos + 5.7735} cy={ypos + 5} r={2.5} fill='#FFFFFF' onClick={onClick}/>);
        this.paths.push(<text fontSize={2.5} x={xpos + 5.7735} y={ypos+5.3} dominantBaseline="middle" textAnchor="middle">{cityProperties.label}</text>);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderHex(hex: Coordinate, hexType: HexType) {
        if (hexType === HexType.OFFBOARD) {
            return;
        }

        let color: string = "";
        switch (hexType) {
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
                throw new Error("city hexes must be added with renderCityHex");
        }
        let xpos = hex.x * 17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y * 5;

        let onClick: undefined | (() => void);
        if (this.emitOnClick) {
            onClick = () => {
                let event = new CustomEvent('mapClickEvent', {
                    detail: {
                        x: hex.x,
                        y: hex.y,
                    }
                });
                document.dispatchEvent(event);
            }
        }

        let points = `${xpos},${ypos + 5} ${xpos + 2.887},${ypos} ${xpos + 8.661},${ypos} ${xpos + 11.547},${ypos + 5} ${xpos + 8.661},${ypos + 10} ${xpos + 2.887},${ypos + 10}`
        this.paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

        if (hexType === HexType.TOWN) {
            this.paths.push(<circle cx={xpos + 5.7735} cy={ypos + 5} r={2.5} fill='#FFFFFF' onClick={onClick}/>);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderSpecialCost(hex: Coordinate, cost: number) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        // const s = 5.4, a=3.118
        // [[0, 2.7], [1.559, 0], [4.677,0], [6.236, 2.7], [4.677, 5.4], [1.559, 5.4]]
        // dx=(11.547-6.236)/2, dy=(10-5.4)/2,
        // dx=2.6555, dy=2.3
        // [[ 2.6555, 5 ], [ 4.2145, 2.3 ], [ 7.3325, 2.3 ], [ 8.8915, 5 ], [ 7.3325, 7.7 ], [ 4.2145, 7.7 ]]
        let points = `${xpos + 2.6555},${ypos + 5} ${xpos + 4.2145},${ypos + 2.3} ${xpos + 7.3325},${ypos + 2.3} ${xpos + 8.8915},${ypos + 5} ${xpos + 7.3325},${ypos + 7.7} ${xpos + 4.2145},${ypos + 7.7}`
        this.paths.push(<polygon fill='#cfddbb' points={points} />);
        this.paths.push(<text fill='#b63421' fontSize={2.5} x={xpos + 5.7735} y={ypos+5.3} dominantBaseline="middle" textAnchor="middle">${cost}</text>);
    }

    public renderTownTrack(hex: Coordinate, direction: Direction, playerColor: PlayerColor|undefined) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        let offset = hexEdgeOffset(direction);
        this.paths.push(<line stroke={playerColorToHtml(playerColor)} strokeWidth={1} x1={xpos+5.7735} y1={ypos+5} x2={xpos+offset.dx} y2={ypos+offset.dy} />);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderTrack(hex: Coordinate, left: Direction, right: Direction, playerColor: PlayerColor|undefined) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        let htmlColor = playerColorToHtml(playerColor);

        let leftOffset = hexEdgeOffset(left);
        let rightOffset = hexEdgeOffset(right);

        let edgeDelta = Math.abs(right - left);
        if (edgeDelta === 1 || edgeDelta === 5) {
            // Tight curve
            let controlX = 5.7735 + (leftOffset.dx - 5.7735)/4 + (rightOffset.dx - 5.7735)/4
            let controlY = 5 + (leftOffset.dy - 5)/4 + (rightOffset.dy - 5)/4

            this.paths.push(<path stroke={htmlColor} strokeWidth={1} fill="none" d={`M ${xpos+leftOffset.dx} ${ypos+leftOffset.dy} Q ${xpos+controlX} ${ypos+controlY} ${xpos+rightOffset.dx} ${ypos+rightOffset.dy}`} />);
        } else if (edgeDelta === 2 || edgeDelta === 4) {
            // Gentle curve
            this.paths.push(<path stroke={htmlColor} strokeWidth={1} fill="none" d={`M ${xpos+leftOffset.dx} ${ypos+leftOffset.dy} Q ${xpos+5.7735} ${ypos+5} ${xpos+rightOffset.dx} ${ypos+rightOffset.dy}`} />);
        } else {
            // Straight
            this.paths.push(<line stroke={htmlColor} strokeWidth={1} x1={xpos+leftOffset.dx} y1={ypos+leftOffset.dy} x2={xpos+rightOffset.dx} y2={ypos+rightOffset.dy} />);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderCubes(hex: Coordinate, cubes: Color[]) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        // Center the xpos
        xpos += (11.547-cubes.length*2.5+0.5)/2;

        for (let i = 0; i < cubes.length; i++) {
            let cube = cubes[i];
            if (cube === Color.NONE) {
                continue;
            }

            let onClick: undefined | (() => void);
            if (this.emitOnClick) {
                onClick = () => {
                    let event = new CustomEvent('cubeClickEvent', {
                        detail: {
                            color: cubes[i],
                            x: hex.x,
                            y: hex.y,
                        }
                    });
                    document.dispatchEvent(event);
                }
            }

            let points = `${xpos+i*2.5},${ypos+0.5} ${xpos+2+i*2.5},${ypos+0.5} ${xpos+2+i*2.5},${ypos+2.5} ${xpos+i*2.5},${ypos+2.5}`
            this.paths.push(<polygon stroke='#222222' strokeWidth={0.25} fill={colorToHtml(cube)} points={points} filter={`url(#${this.filterId})`} onClick={onClick}/>);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderActiveCube(hex: Coordinate, cube: Color) {
        if (cube === Color.NONE) {
            return;
        }

        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        // Center the xpos,ypos based on a 3.5x3.5 size
        xpos += 4.0235;
        ypos += 3.25;

        let points = `${xpos},${ypos} ${xpos+2.5},${ypos} ${xpos+2.5},${ypos+2.5} ${xpos},${ypos+2.5}`
        this.paths.push(<polygon stroke='#FFFF00' strokeWidth={0.5} fill={colorToHtml(cube)} points={points} filter={`url(#${this.filterId})`}/>);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderArrow(hex: Coordinate, direction: Direction, color: PlayerColor|undefined) {
        let htmlColor = playerColorToHtml(color);

        // Big hex
        // W        NW       NE        E           SE          SW
        // const s = 10, a=5.774;
        // [[0, s/2], [a/2, 0], [a*1.5,0], [a*2, s/2], [a*1.5, s], [a/2, s]]

        // Smaller hex
        // const s = 8, a=4.619

        // offset [1, 1.155]

        // W                NW                  NE                 E                  SE                 SW
        // [ [ 1, 5.155 ],  [ 3.3095, 1.155 ],  [ 7.9285, 1.155 ], [ 10.238, 5.155 ], [ 7.9285, 9.155 ], [ 3.3095, 9.155 ] ]


        // /     |
        // ------|  2.155
        // 4.619
        // 0.436532338 rad
        // a=2.155
        // b=2.3095
        // alpha=alpha=Math.atan(a/b)
        // beta=Math.atan((p[0][0]-p[1][0])/(p[0][1]-p[1][1]))
        // gamma=pi/2-alpha-beta+pi/2
        // x=h*Math.cos(gamma) + p[0][0]
        // y=h*Math.sin(gamma) + p[0][1]

        let points: number[][];
        switch (direction) {
            case Direction.NORTH:
                points = [[3.3095, 9.155], [7.9285, 9.155], [5.774, 7]]
                break;
            case Direction.NORTH_EAST:
                points = [[1, 5.155], [3.3095, 9.155], [4.02, 6.077]]
                break;
            case Direction.SOUTH_EAST:
                points = [[1, 5.155], [3.3095, 1.155], [4.02, 4.232]]
                break;
            case Direction.SOUTH:
                points = [[3.3095, 1.155], [7.9285, 1.155], [5.774, 3.31]]
                break;
            case Direction.SOUTH_WEST:
                points = [[7.9285, 1.155], [10.238, 5.155], [7.006, 4.175]]
                break;
            case Direction.NORTH_WEST:
                points = [[10.238, 5.155], [7.9285, 9.155], [7.006, 5.866]]
                break;
            default:
                throw new Error("Unhandled direction: " + direction);
        }

        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;
        for (let point of points) {
            point[0] += xpos;
            point[1] += ypos;
        }

        let onClick: undefined | (() => void);
        if (this.emitOnClick) {
            onClick = () => {
                let event = new CustomEvent('arrowClickEvent', {
                    detail: {
                        hex: hex,
                        direction: direction,
                    }
                });
                document.dispatchEvent(event);
            }
        }

        this.paths.push(<polygon stroke='#FFFFFF' strokeWidth={0.5} fill={htmlColor} onClick={onClick} points={points.map(p => p.join(",")).join(" ")} filter={`url(#${this.filterId})`} />);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public render(): ReactNode {
        let width = this.width+1;
        let height = this.height+1;

        let pixelWidth;
        if (width == 1) {
            pixelWidth = 11.547;
        } else {
            pixelWidth = width * 17.321 + 8.661;
        }
        let pixelHeight;
        if (height == 1) {
            pixelHeight = 10;
        } else {
            pixelHeight = (height+1)/2 * 10;
            if ((height % 2) == 0) {
                pixelHeight += 5;
            }
        }

        return <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox={`0 0 ${pixelWidth} ${pixelHeight}`}>
            <defs>
                <filter id={'' + this.filterId} width="2.5" height="2.5">
                    <feOffset in="SourceAlpha" dx="0.5" dy="0.5"/>
                    <feGaussianBlur stdDeviation="0.25"/>
                    <feBlend in="SourceGraphic" in2="blurOut"/>
                </filter>
            </defs>
            {this.paths}
        </svg>;
    }
}
