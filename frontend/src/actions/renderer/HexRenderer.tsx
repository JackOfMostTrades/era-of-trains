import {ReactNode, useId} from "react";
import {Color, Coordinate, Direction, PlayerColor} from "../../api/api.ts";
import {CityProperties, HexType} from "../../maps";
import {UserSession} from "../../UserSessionContext.tsx";

export function colorToHtml(color: Color, userSession: UserSession|undefined): string {
    if (color == Color.NONE) {
        return '#ffffff';
    }

    let customColor = userSession?.userProfile?.customColors?.goodsColors?.[color];
    if (customColor) {
        return customColor;
    }

    switch (color) {
        case Color.BLACK: return "#444444";
        case Color.BLUE: return '#00a2d3';
        case Color.RED: return '#d41e2d';
        case Color.PURPLE: return '#8d5a95';
        case Color.YELLOW: return '#f5ac11';
        case Color.WHITE: return '#e9dcc9';
    }
}

export function playerColorToHtml(color: PlayerColor|undefined, userSession: UserSession|undefined): string {
    if (color === undefined) {
        return '#111111';
    }

    let customColor = userSession?.userProfile?.customColors?.playerColors?.[color];
    if (customColor) {
        return customColor;
    }

    switch (color) {
        case PlayerColor.BLUE: return '#035f70';
        case PlayerColor.GREEN: return "#a7ed36";
        case PlayerColor.YELLOW: return '#f3db70';
        case PlayerColor.PINK: return '#bf88a9';
        case PlayerColor.GRAY: return '#495049';
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

function HexName({centerX, centerY, name}: {centerX: number, centerY: number, name: string}) {
    const pathId = useId();

    return <>
        <path id={"curve" + pathId} stroke="none" fill="none" d={`M ${centerX - 3.6} ${centerY} a 2.9 2.6 0 0 0 7.2 0`}/>
        <text fontSize={1.4} fill="white" dominantBaseline="middle" textAnchor="middle">
            <textPath xlinkHref={"#curve" + pathId} startOffset="50%">
                {name}
            </textPath>
        </text>
    </>
}

export class HexRenderer {
    private width: number;
    private height: number;
    private paths: ReactNode[];
    private emitOnClick: boolean;
    private userSession: UserSession|undefined;
    private shouldRenderGridLegend: boolean;
    private filterId: number;
    private static nextFilterId: number = 0;

    constructor(emitOnClick: boolean, shouldRenderGridLegend: boolean, userSession: UserSession|undefined) {
        this.width = -1;
        this.height = -1;
        this.paths = [];
        this.emitOnClick = emitOnClick;
        this.shouldRenderGridLegend = shouldRenderGridLegend;
        this.userSession = userSession;
        this.filterId = HexRenderer.nextFilterId;
        HexRenderer.nextFilterId += 1;
    }
    
    private getHexXY(hex: Coordinate): {x: number, y: number} {
        let x = hex.x*17.321;
        if ((hex.y % 2) === 1) {
            x += 8.661;
        }
        let y = hex.y*5;
        return {x: x, y: y};
    }

    private getHexCenter(hex: Coordinate): {x: number, y: number} {
        let pos = this.getHexXY(hex);

        // Offset to the center of the hex
        pos.x += 5.7735;
        pos.y += 5;

        return pos;
    }

    public renderCityHex(hex: Coordinate, cityProperties: CityProperties, locationName: string|undefined) {
        let pos = this.getHexXY(hex);

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

        let points = `${pos.x},${pos.y+5} ${pos.x+2.887},${pos.y} ${pos.x+8.661},${pos.y} ${pos.x+11.547},${pos.y+5} ${pos.x+8.661},${pos.y+10} ${pos.x+2.887},${pos.y+10}`
        this.paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

        let cityColor: string;
        if (cityProperties.color === Color.NONE) {
            cityColor = "#cccccc";
        } else {
            cityColor = colorToHtml(cityProperties.color, this.userSession);
        }
        let strokeColor: string = '#222222';
        if (cityProperties.darkCity) {
            strokeColor = '#ffffff';
        }
        points = `${pos.x + 0.8},${pos.y + 5} ${pos.x + 3.225},${pos.y + 0.8} ${pos.x + 8.075},${pos.y + 0.8} ${pos.x + 10.747},${pos.y + 5} ${pos.x + 8.075},${pos.y + 9.2} ${pos.x + 3.225},${pos.y + 9.2}`
        this.paths.push(<polygon stroke={strokeColor} strokeWidth={0.2} fill={cityColor} points={points} onClick={onClick}/>);
        this.paths.push(<circle cx={pos.x + 5.7735} cy={pos.y + 5} r={2.2} fill='#FFFFFF' onClick={onClick}/>);
        this.paths.push(<text fontSize={2.5} x={pos.x + 5.7735} y={pos.y+5.3} dominantBaseline="middle" textAnchor="middle">{cityProperties.label}</text>);
        if (locationName) {
            let center = this.getHexCenter(hex);
            this.paths.push(<HexName centerX={center.x} centerY={center.y} name={locationName} />);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderHex(hex: Coordinate, hexType: HexType, locationName: string|undefined) {
        if (hexType === HexType.OFFBOARD) {
            return;
        }

        let color: string = "";
        switch (hexType) {
            case HexType.WATER:
                color = '#579ba8';
                break;
            case HexType.RIVER:
                color = '#8eb572';
                break;
            case HexType.HILLS:
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
                throw new Error(`city hex (${hex.x}, ${hex.y}) must be added with renderCityHex`);
        }
        let pos = this.getHexXY(hex);

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

        let points = `${pos.x},${pos.y + 5} ${pos.x + 2.887},${pos.y} ${pos.x + 8.661},${pos.y} ${pos.x + 11.547},${pos.y + 5} ${pos.x + 8.661},${pos.y + 10} ${pos.x + 2.887},${pos.y + 10}`
        this.paths.push(<polygon stroke='#000000' strokeWidth={0.1} fill={color} points={points} onClick={onClick}/>);

        if (hexType === HexType.TOWN) {
            this.paths.push(<circle cx={pos.x + 5.7735} cy={pos.y + 5} r={2.2} fill='#FFFFFF' onClick={onClick}/>);
            if (locationName) {
                let center = this.getHexCenter(hex);
                this.paths.push(<HexName centerX={center.x} centerY={center.y} name={locationName} />);
            }
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderSpecialCost(hex: Coordinate, cost: number) {
        let pos = this.getHexXY(hex);

        // const s = 5.4, a=3.118
        // [[0, 2.7], [1.559, 0], [4.677,0], [6.236, 2.7], [4.677, 5.4], [1.559, 5.4]]
        // dx=(11.547-6.236)/2, dy=(10-5.4)/2,
        // dx=2.6555, dy=2.3
        // [[ 2.6555, 5 ], [ 4.2145, 2.3 ], [ 7.3325, 2.3 ], [ 8.8915, 5 ], [ 7.3325, 7.7 ], [ 4.2145, 7.7 ]]
        let points = `${pos.x + 2.6555},${pos.y + 5} ${pos.x + 4.2145},${pos.y + 2.3} ${pos.x + 7.3325},${pos.y + 2.3} ${pos.x + 8.8915},${pos.y + 5} ${pos.x + 7.3325},${pos.y + 7.7} ${pos.x + 4.2145},${pos.y + 7.7}`
        this.paths.push(<polygon fill='#cfddbb' points={points} />);
        this.paths.push(<text fill='#b63421' fontSize={2.5} x={pos.x + 5.7735} y={pos.y+5.3} dominantBaseline="middle" textAnchor="middle">${cost}</text>);
    }

    public renderTownTrack(hex: Coordinate, direction: Direction, playerColor: PlayerColor|undefined) {
        let pos = this.getHexXY(hex);

        let htmlColor = playerColorToHtml(playerColor, this.userSession);
        let dashArray: string|undefined;
        if (playerColor === undefined) {
            dashArray = '1 0.25';
        }

        let offset = hexEdgeOffset(direction);
        this.paths.push(<line stroke={htmlColor} strokeWidth={1} fill="none" strokeDasharray={dashArray} x1={pos.x+5.7735} y1={pos.y+5} x2={pos.x+offset.dx} y2={pos.y+offset.dy} />);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderTrack(hex: Coordinate, left: Direction, right: Direction, playerColor: PlayerColor|undefined) {
        let pos = this.getHexXY(hex);

        let htmlColor = playerColorToHtml(playerColor, this.userSession);
        let dashArray: string|undefined;
        if (playerColor === undefined) {
            dashArray = '1 0.25';
        }

        let leftOffset = hexEdgeOffset(left);
        let rightOffset = hexEdgeOffset(right);

        let edgeDelta = Math.abs(right - left);
        if (edgeDelta === 1 || edgeDelta === 5) {
            // Tight curve
            let controlX = 5.7735 + (leftOffset.dx - 5.7735)/4 + (rightOffset.dx - 5.7735)/4
            let controlY = 5 + (leftOffset.dy - 5)/4 + (rightOffset.dy - 5)/4

            this.paths.push(<path stroke={htmlColor} strokeWidth={1} fill="none" strokeDasharray={dashArray} d={`M ${pos.x+leftOffset.dx} ${pos.y+leftOffset.dy} Q ${pos.x+controlX} ${pos.y+controlY} ${pos.x+rightOffset.dx} ${pos.y+rightOffset.dy}`} />);
        } else if (edgeDelta === 2 || edgeDelta === 4) {
            // Gentle curve
            this.paths.push(<path stroke={htmlColor} strokeWidth={1} fill="none" strokeDasharray={dashArray} d={`M ${pos.x+leftOffset.dx} ${pos.y+leftOffset.dy} Q ${pos.x+5.7735} ${pos.y+5} ${pos.x+rightOffset.dx} ${pos.y+rightOffset.dy}`} />);
        } else {
            // Straight
            this.paths.push(<line stroke={htmlColor} strokeWidth={1} fill="none" strokeDasharray={dashArray} x1={pos.x+leftOffset.dx} y1={pos.y+leftOffset.dy} x2={pos.x+rightOffset.dx} y2={pos.y+rightOffset.dy} />);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderCubes(hex: Coordinate, cubes: Color[]) {
        let pos = this.getHexXY(hex);

        // Center the pos.x
        pos.x += (11.547-cubes.length*2.5+0.5)/2;

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

            let points = `${pos.x+i*2.5},${pos.y+0.5} ${pos.x+2+i*2.5},${pos.y+0.5} ${pos.x+2+i*2.5},${pos.y+2.5} ${pos.x+i*2.5},${pos.y+2.5}`
            this.paths.push(<polygon stroke='#222222' strokeWidth={0.25} fill={colorToHtml(cube, this.userSession)} points={points} filter={`url(#${this.filterId})`} onClick={onClick}/>);
        }

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderActiveCube(hex: Coordinate, cube: Color, moveAlong: Coordinate[]|undefined) {
        if (cube === Color.NONE) {
            return;
        }

        let cubeCoord = this.getHexCenter(hex);

        let animate: ReactNode
        if (moveAlong && moveAlong.length > 0) {
            let start = this.getHexCenter(moveAlong[0]);
            let path = `M${start.x-cubeCoord.x},${start.y-cubeCoord.y}`
            for (let i = 1; i < moveAlong.length; i++) {
                let point = this.getHexCenter(moveAlong[i]);
                path += ` L${point.x-cubeCoord.x},${point.y-cubeCoord.y}`
            }
            path += " L0,0"

            animate = <animateMotion
                ref={ref => {
                    if (ref) {
                        (ref as SVGAnimateMotionElement).beginElement();
                    }
                }}
                begin="indefinite" dur={(100*(moveAlong.length+1)) + "ms"} path={path} />
        }
        let points = `${cubeCoord.x-1.25},${cubeCoord.y-1.25} ${cubeCoord.x+1.25},${cubeCoord.y-1.25} ${cubeCoord.x+1.25},${cubeCoord.y+1.25} ${cubeCoord.x-1.25},${cubeCoord.y+1.25}`
        this.paths.push(<polygon stroke='#FFFF00' strokeWidth={0.5} fill={colorToHtml(cube, this.userSession)} points={points} filter={`url(#${this.filterId})`}>{animate}</polygon>);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderArrow(hex: Coordinate, direction: Direction, color: PlayerColor|undefined) {
        let htmlColor = playerColorToHtml(color, this.userSession);

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

        let pos = this.getHexXY(hex);
        for (let point of points) {
            point[0] += pos.x;
            point[1] += pos.y;
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

    public highlightHex(hex: Coordinate) {
        let pos = this.getHexXY(hex);

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

        let points = `${pos.x},${pos.y + 5} ${pos.x + 2.887},${pos.y} ${pos.x + 8.661},${pos.y} ${pos.x + 11.547},${pos.y + 5} ${pos.x + 8.661},${pos.y + 10} ${pos.x + 2.887},${pos.y + 10}`
        this.paths.push(<polygon strokeWidth={0} fill="#ffffff" opacity={0.3} points={points} onClick={onClick}/>);

        this.width = Math.max(this.width, hex.x);
        this.height = Math.max(this.height, hex.y);
    }

    public renderTeleportLink(id: {hex: Coordinate, direction: Direction}, hex: Coordinate, offset: Direction|-1, fillColor: PlayerColor|undefined, cost: number|undefined) {

        let pos: {x: number, y: number};
        if (offset === -1) {
            pos = this.getHexCenter(hex);
        } else {
            pos = this.getHexXY(hex);
            let edge = hexEdgeOffset(offset);
            pos.x += edge.dx;
            pos.y += edge.dy;
        }

        let fill: string;
        if (fillColor === undefined) {
            fill = '#ffffff';
        } else {
            fill = playerColorToHtml(fillColor, this.userSession);
        }

        let onClick: (() => void) | undefined;
        if (this.emitOnClick) {
            onClick = () => {
                let event = new CustomEvent('teleportLinkClickEvent', {
                    detail: {
                        hex: id.hex,
                        direction: id.direction,
                    }
                });
                document.dispatchEvent(event);
            }
        }

        this.paths.push(<circle onClick={onClick} stroke='#000000' strokeWidth={0.25} cx={pos.x} cy={pos.y} r={2} fill={fill} />);
        if (cost !== undefined) {
            this.paths.push(<text fill='#b63421' fontSize={1.5} x={pos.x} y={pos.y+0.125} dominantBaseline="middle" textAnchor="middle">${cost}</text>);
        }
    }

    public renderLayer(node: ReactNode) {
        this.paths.push(node);
    }

    private renderGridLegend(): ReactNode {
        if (!this.shouldRenderGridLegend) {
            return null;
        }
        let nodes = [];
        for (let x = 0; x < (this.width+1)*2; x++) {
            let xPos = 7.5 + x * 8.661 + 11.547/2;
            nodes.push(<text key={"x" + x} fontSize={2.5} x={xPos} y={2.5} dominantBaseline="middle" textAnchor="middle">{'' + (x+1)}</text>);
        }
        for (let y = 0; y < this.height+1; y++) {
            let yPos = 5 + y*5 + 5;
            let label = String.fromCharCode('A'.charCodeAt(0) + (y%26));
            if (y >= 26) {
                label = 'A' + label;
            }
            nodes.push(<text key={"y" + y} fontSize={2.5} x={y%2 == 0 ? 2.5 : 5.0} y={yPos} dominantBaseline="middle" textAnchor="middle">{label}</text>);
        }
        return nodes;
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
        if (this.shouldRenderGridLegend) {
            pixelWidth += 7.5;
            pixelHeight += 5;
        }

        let paths = <>{this.paths}</>;
        if (this.shouldRenderGridLegend) {
            paths = <g transform="translate(7.5,5)">
                {paths}
            </g>
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
            {this.renderGridLegend()}
            {paths}
        </svg>;
    }
}
