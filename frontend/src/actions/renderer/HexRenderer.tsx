import {Color, Coordinate, Direction} from "../../api/api.ts";
import {HexType} from "../../map.ts";
import {ReactNode} from "react";

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

export interface CityState {
    label: string
    color: Color
    darkCity: boolean
}

export function urbCityState(n: number): CityState {
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

    constructor(emitOnClock: boolean) {
        this.width = -1;
        this.height = -1;
        this.paths = [];
        this.emitOnClick = emitOnClock;
    }

    public renderCityHex(hex: Coordinate, cityState: CityState) {
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
        if (cityState.darkCity) {
            color = '#222222';
        }

        let points = `${xpos},${ypos+5} ${xpos+2.887},${ypos} ${xpos+8.661},${ypos} ${xpos+11.547},${ypos+5} ${xpos+8.661},${ypos+10} ${xpos+2.887},${ypos+10}`
        this.paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

        let cityColor = colorToHtml(cityState.color);
        let strokeColor: string = '#222222';
        if (cityState.darkCity) {
            strokeColor = '#ffffff';
        }
        points = `${xpos + 0.8},${ypos + 5} ${xpos + 3.225},${ypos + 0.8} ${xpos + 8.075},${ypos + 0.8} ${xpos + 10.747},${ypos + 5} ${xpos + 8.075},${ypos + 9.2} ${xpos + 3.225},${ypos + 9.2}`
        this.paths.push(<polygon stroke={strokeColor} strokeWidth={0.2} fill={cityColor} points={points} onClick={onClick}/>);
        this.paths.push(<circle cx={xpos + 5.7735} cy={ypos + 5} r={2.5} fill='#FFFFFF' onClick={onClick}/>);
        this.paths.push(<text fontSize={2.5} x={xpos + 5.7735} y={ypos+5.3} dominantBaseline="middle" textAnchor="middle">{cityState.label}</text>);

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

    public renderTownTrack(hex: Coordinate, direction: Direction, player: string) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        let offset = hexEdgeOffset(direction);
        // FIXME: render with player color
        this.paths.push(<line stroke='#222222' strokeWidth={1} x1={xpos+5.7735} y1={ypos+5} x2={xpos+offset.dx} y2={ypos+offset.dy} />);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderTrack(hex: Coordinate, left: Direction, right: Direction, player: string) {
        let xpos = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            xpos += 8.661;
        }
        let ypos = hex.y*5;

        let leftOffset = hexEdgeOffset(left);
        let rightOffset = hexEdgeOffset(right);
        // FIXME: render with player color
        this.paths.push(<line stroke='#222222' strokeWidth={1} x1={xpos+leftOffset.dx} y1={ypos+leftOffset.dy} x2={xpos+rightOffset.dx} y2={ypos+rightOffset.dy} />);

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
            this.paths.push(<polygon stroke='#222222' strokeWidth={0.25} fill={colorToHtml(cube)} points={points} filter="url(#cube-shadow)" onClick={onClick}/>);
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
        this.paths.push(<polygon stroke='#FFFF00' strokeWidth={0.5} fill={colorToHtml(cube)} points={points} filter="url(#cube-shadow)"/>);

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
            pixelHeight = height * 10;
            if ((pixelHeight % 2) === 0) {
                pixelHeight += 5;
            }
        }

        return <svg
            xmlns="http://www.w3.org/2000/svg"
            width={pixelWidth * 6}
            height={pixelHeight * 6}
            viewBox={`0 0 ${pixelWidth} ${pixelHeight}`}>
            <defs>
                <filter id="cube-shadow" width="2.5" height="2.5">
                    <feOffset in="SourceAlpha" dx="0.5" dy="0.5"/>
                    <feGaussianBlur stdDeviation="0.25"/>
                    <feBlend in="SourceGraphic" in2="blurOut"/>
                </filter>
            </defs>
            {this.paths}
        </svg>;
    }
}
