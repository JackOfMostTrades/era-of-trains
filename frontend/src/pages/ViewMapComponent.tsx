import {BuildAction, Color, Coordinate, Direction, GameState, PlayerColor} from "../api/api.ts";
import {ReactNode, useContext, useEffect, useState} from "react";
import {CityProperties, GameMap, HexType} from "../maps";
import {HexRenderer, urbCityProperties} from "../actions/renderer/HexRenderer.tsx";
import {applyDirection, applyMapDirection, oppositeDirection} from "../util.ts";
import {Step as MoveGoodsStep} from "../actions/MoveGoodsActionSelector.tsx";
import UserSessionContext, {UserSession} from "../UserSessionContext.tsx";
import {renderTrackTile} from "../actions/TrackSelector.tsx";
import {MapTileState, Rotation} from "../game/map_state.ts";

function getCityProperties(gameState: GameState|undefined, map: GameMap, coordinate: Coordinate): CityProperties|undefined {
    if (gameState && gameState.urbanizations) {
        for (let urb of gameState.urbanizations) {
            if (urb.hex.x === coordinate.x && urb.hex.y === coordinate.y) {
                return urbCityProperties(urb.city);
            }
        }
    }

    return map.getCityProperties(gameState, coordinate);
}

class RenderMapBuilder {
    private gameState: GameState|undefined;
    private map: GameMap;
    public hexRenderer: HexRenderer;

    constructor(gameState: GameState|undefined, map: GameMap, userSession: UserSession|undefined) {
        this.gameState = gameState;
        this.map = map;
        this.hexRenderer = new HexRenderer(true, true, userSession);
    }

    public renderCityHex(hex: Coordinate, cityProperties: CityProperties) {
        this.hexRenderer.renderCityHex(hex, cityProperties);
    }

    public renderHex(hex: Coordinate) {
        let cityProperties: CityProperties|undefined = getCityProperties(this.gameState, this.map, hex);
        if (cityProperties) {
            this.hexRenderer.renderCityHex(hex, cityProperties);
        } else {
            this.hexRenderer.renderHex(hex, this.map.getHexType(hex));
        }
    }

    public renderLayer(node: ReactNode) {
        this.hexRenderer.renderLayer(node);
    }

    public renderEmptyTeleportLink(id: {hex: Coordinate, direction: Direction}, hex: Coordinate, offset: Direction|-1, cost: number) {
        this.hexRenderer.renderTeleportLink(id, hex, offset, undefined, cost);
    }

    public renderOccupiedTeleportLink(id: {hex: Coordinate, direction: Direction}, hex: Coordinate, offset: Direction|-1, playerColor: PlayerColor) {
        this.hexRenderer.renderTeleportLink(id, hex, offset, playerColor, undefined);
    }

    public renderSpecialCost(hex: Coordinate, cost: number) {
        this.hexRenderer.renderSpecialCost(hex, cost);
    }

    public renderTownTrack(hex: Coordinate, direction: Direction, player: string) {
        this.hexRenderer.renderTownTrack(hex, direction, this.gameState?.playerColor[player]);
    }

    public renderTrack(hex: Coordinate, left: Direction, right: Direction, player: string) {
        this.hexRenderer.renderTrack(hex, left, right, this.gameState?.playerColor[player]);
    }

    public renderCubes(hex: Coordinate, cubes: Color[]) {
        this.hexRenderer.renderCubes(hex, cubes);
    }

    public renderActiveCube(hex: Coordinate, cube: Color, moveAlong: Coordinate[]|undefined) {
        this.hexRenderer.renderActiveCube(hex, cube, moveAlong);
    }

    public renderArrow(hex: Coordinate, direction: Direction, color: PlayerColor|undefined) {
        this.hexRenderer.renderArrow(hex, direction, color);
    }

    public render(): ReactNode {
        return this.hexRenderer.render();
    }
}

function ViewMapComponent({gameState, activePlayer, map}: {gameState: GameState|undefined, activePlayer: string, map: GameMap}) {
    let userSession = useContext(UserSessionContext);
    let [pendingBuildAction, setPendingBuildAction] = useState<BuildAction|undefined>(undefined);
    let [pendingMoveGoods, setPendingMoveGoods] = useState<MoveGoodsStep|undefined>(undefined)

    useEffect(() => {
        const handler = (e:CustomEventInit<BuildAction>) => {
            setPendingBuildAction(e.detail);
        };
        document.addEventListener('pendingBuildAction', handler);
        return () => document.removeEventListener('pendingBuildAction', handler);
    }, []);

    useEffect(() => {
        const handler = (e:CustomEventInit<MoveGoodsStep>) => {
            setPendingMoveGoods(e.detail);
        };
        document.addEventListener('pendingMoveGoods', handler);
        return () => document.removeEventListener('pendingMoveGoods', handler);
    }, []);

    let renderer = new RenderMapBuilder(gameState, map, userSession);

    for (let y = 0; y < map.getHeight(); y++) {
        for (let x = 0; x < map.getWidth(); x++) {
            let hex: Coordinate = {x: x, y: y};
            renderer.renderHex(hex);

            let specialCost = map.getSpecialTrackPricing(hex);
            if (specialCost !== undefined) {
                renderer.renderSpecialCost(hex, specialCost);
            }
        }
    }
    renderer.renderLayer(map.getRiverLayer());

    for (let teleportLink of map.getTeleportLinks(gameState, pendingBuildAction)) {
        let owner: string|undefined;
        if (gameState && gameState.links) {
            for (let playerLink of gameState.links) {
                let hex = playerLink.sourceHex;
                for (let step of playerLink.steps) {
                    if ((hex.x === teleportLink.left.hex.x && hex.y === teleportLink.left.hex.y && step === teleportLink.left.direction) ||
                            (hex.x === teleportLink.right.hex.x && hex.y === teleportLink.right.hex.y && step === teleportLink.right.direction)) {
                        owner = playerLink.owner;
                        break;
                    }
                    hex = applyMapDirection(map, gameState, pendingBuildAction, hex, step);
                }
                if (owner) {
                    break;
                }
            }
        }
        if (pendingBuildAction && pendingBuildAction.teleportLinkPlacements) {
            for (let playerLink of pendingBuildAction.teleportLinkPlacements) {
                let hex = playerLink.hex;
                let step = playerLink.track;
                if ((hex.x === teleportLink.left.hex.x && hex.y === teleportLink.left.hex.y && step === teleportLink.left.direction) ||
                    (hex.x === teleportLink.right.hex.x && hex.y === teleportLink.right.hex.y && step === teleportLink.right.direction)) {
                    owner = activePlayer;
                    break;
                }
            }
        }
        if (owner) {
            renderer.renderOccupiedTeleportLink({
                hex: teleportLink.left.hex,
                direction: teleportLink.left.direction,
            }, teleportLink.costLocation, teleportLink.costLocationEdge, gameState?.playerColor[owner] as PlayerColor);
        } else {
            renderer.renderEmptyTeleportLink({
                hex: teleportLink.left.hex,
                direction: teleportLink.left.direction,
            }, teleportLink.costLocation, teleportLink.costLocationEdge, teleportLink.cost);
        }
    }

    if (gameState) {
        // Render links
        if (gameState.links) {
            for (let link of gameState.links) {
                let hex = link.sourceHex;
                let cityProperties = getCityProperties(gameState, map, hex);
                if (!cityProperties) {
                    if (map.getHexType(hex) !== HexType.TOWN) {
                        throw new Error("link started at non-city and non-town");
                    }
                    renderer.renderTownTrack(hex, link.steps[0], link.owner);
                }

                for (let i = 1; i < link.steps.length; i++) {
                    hex = applyMapDirection(map, gameState, pendingBuildAction, hex, link.steps[i-1]);

                    // If there is a track placement on this hex, skip rendering here
                    let hasTilePlacement = false;
                    if (pendingBuildAction && pendingBuildAction.trackPlacements) {
                        for (let trackPlacement of pendingBuildAction.trackPlacements) {
                            if (trackPlacement.hex.x === hex.x && trackPlacement.hex.y === hex.y) {
                                hasTilePlacement = true;
                                break;
                            }
                        }
                    }
                    if (hasTilePlacement) {
                        continue;
                    }

                    let cityProperties = getCityProperties(gameState, map, hex);
                    if (!cityProperties) {
                        let left = oppositeDirection(link.steps[i-1]);
                        let right = link.steps[i];

                        if (map.getHexType(hex) === HexType.TOWN) {
                            renderer.renderTownTrack(hex, left, link.owner);
                        } else {
                            renderer.renderTrack(hex, left, right, link.owner);
                        }
                    }
                }

                // Render the last step in a completed link to a town
                hex = applyMapDirection(map, gameState, pendingBuildAction, hex, link.steps[link.steps.length-1]);
                if (link.complete && map.getHexType(hex) === HexType.TOWN) {
                    let cityProperties = getCityProperties(gameState, map, hex);
                    if (!cityProperties) {
                        renderer.renderTownTrack(hex, oppositeDirection(link.steps[link.steps.length - 1]), link.owner);
                    }
                }
            }
        }

        // Render pending build action
        if (pendingBuildAction) {
            let mapTileState = new MapTileState(map, gameState);

            if (pendingBuildAction.urbanization) {
                renderer.renderCityHex(pendingBuildAction.urbanization.hex, urbCityProperties(pendingBuildAction.urbanization.city));
            }
            for (let townPlacement of pendingBuildAction.townPlacements) {
                renderer.renderTownTrack(townPlacement.hex, townPlacement.track, activePlayer);
            }
            for (let trackPlacement of pendingBuildAction.trackPlacements) {
                renderTrackTile(mapTileState.getTileState(trackPlacement.hex).routes,
                        gameState,
                        activePlayer,
                        trackPlacement.hex,
                        trackPlacement.tile,
                        trackPlacement.rotation as Rotation,
                        renderer.hexRenderer);
            }
        }

        if (gameState.cubes) {
            let coordByKey: { [key: string]: Coordinate } = {};
            let cubesByKey: { [key: string]: Color[] } = {};
            for (let cube of gameState.cubes) {
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

                // If we are moving a cube, skip rendering it (by setting color to none, so we leave a hole in the rendering)
                if (pendingMoveGoods &&
                    pendingMoveGoods.selectedColor !== Color.NONE &&
                    pendingMoveGoods.selectedOrigin &&
                    pendingMoveGoods.selectedOrigin.x === hex.x && pendingMoveGoods.selectedOrigin.y === hex.y) {

                    for (let i = 0; i < cubes.length; i++) {
                        if (cubes[i] === pendingMoveGoods.selectedColor) {
                            cubes[i] = Color.NONE;
                            break;
                        }
                    }
                }

                renderer.renderCubes(hex, cubes);
            }
        }

        if (pendingMoveGoods && pendingMoveGoods.selectedColor !== Color.NONE
                && pendingMoveGoods.selectedColor !== undefined
                && pendingMoveGoods.currentCubePosition) {
            renderer.renderActiveCube(pendingMoveGoods.currentCubePosition, pendingMoveGoods.selectedColor,
                pendingMoveGoods.moveAlong);
            if (pendingMoveGoods.nextStepOptions) {
                for (let option of pendingMoveGoods.nextStepOptions) {
                    let hex = applyDirection(pendingMoveGoods.currentCubePosition, option.direction);
                    renderer.renderArrow(hex, option.direction, option.owner);
                }
            }
        }
    }

    return <>
        <div style={{marginTop: '1em'}} />
        {renderer.render()}
    </>
}

export default ViewMapComponent;
