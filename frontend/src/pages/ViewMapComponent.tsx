import {ALL_DIRECTIONS, BuildAction, Color, Coordinate, Direction, PlayerColor, ViewGameResponse} from "../api/api.ts";
import {ReactNode, useEffect, useState} from "react";
import {CityProperties, GameMap, HexType} from "../maps";
import {HexRenderer, urbCityProperties} from "../actions/renderer/HexRenderer.tsx";
import {applyDirection, oppositeDirection} from "../util.ts";
import {Step as MoveGoodsStep} from "../actions/MoveGoodsActionSelector.tsx";

function getCityProperties(game: ViewGameResponse, map: GameMap, coordinate: Coordinate): CityProperties|undefined {
    if (game.gameState && game.gameState.urbanizations) {
        for (let urb of game.gameState.urbanizations) {
            if (urb.hex.x === coordinate.x && urb.hex.y === coordinate.y) {
                return urbCityProperties(urb.city);
            }
        }
    }

    return map.getCityProperties(game.gameState, coordinate);
}

class RenderMapBuilder {
    private game: ViewGameResponse;
    private map: GameMap;
    private hexRenderer: HexRenderer;

    constructor(game: ViewGameResponse, map: GameMap) {
        this.game = game;
        this.map = map;
        this.hexRenderer = new HexRenderer(true);
    }

    public renderCityHex(hex: Coordinate, cityProperties: CityProperties) {
        this.hexRenderer.renderCityHex(hex, cityProperties);
    }

    public renderHex(hex: Coordinate) {
        let cityProperties: CityProperties|undefined = getCityProperties(this.game, this.map, hex);
        if (cityProperties) {
            this.hexRenderer.renderCityHex(hex, cityProperties);
        } else {
            this.hexRenderer.renderHex(hex, this.map.getHexType(hex));
        }
    }

    public renderEmptyInterurbanLink(hex: Coordinate, direction: Direction, cost: number) {
        this.hexRenderer.renderInterurbanLink(hex, direction, undefined, cost);
    }

    public renderOccupiedInterurbanLink(hex: Coordinate, direction: Direction, playerColor: PlayerColor) {
        this.hexRenderer.renderInterurbanLink(hex, direction, playerColor, undefined);
    }

    public renderSpecialCost(hex: Coordinate, cost: number) {
        this.hexRenderer.renderSpecialCost(hex, cost);
    }

    public renderTownTrack(hex: Coordinate, direction: Direction, player: string) {
        this.hexRenderer.renderTownTrack(hex, direction, this.game.gameState?.playerColor[player]);
    }

    public renderTrack(hex: Coordinate, left: Direction, right: Direction, player: string) {
        this.hexRenderer.renderTrack(hex, left, right, this.game.gameState?.playerColor[player]);
    }

    public renderCubes(hex: Coordinate, cubes: Color[]) {
        this.hexRenderer.renderCubes(hex, cubes);
    }

    public renderActiveCube(hex: Coordinate, cube: Color) {
        this.hexRenderer.renderActiveCube(hex, cube);
    }

    public renderArrow(hex: Coordinate, direction: Direction, color: PlayerColor|undefined) {
        this.hexRenderer.renderArrow(hex, direction, color);
    }

    public render(): ReactNode {
        return this.hexRenderer.render();
    }
}

function ViewMapComponent({game, map}: {game: ViewGameResponse, map: GameMap}) {
    let [pendingBuildAction, setPendingBuildAction] = useState<BuildAction|undefined>(undefined);
    let [pendingMoveGoods, setPendingMoveGoods] = useState<MoveGoodsStep|undefined>(undefined)
    let [buildingTrackHex, setBuildingTrackHex] = useState<Coordinate|undefined>(undefined);

    useEffect(() => {
        const handler = (e:CustomEventInit<BuildAction>) => {
            setPendingBuildAction(e.detail);
        };
        document.addEventListener('pendingBuildAction', handler);
        return () => document.removeEventListener('pendingBuildAction', handler);
    }, []);

    useEffect(() => {
        const handler = (e:CustomEventInit<Coordinate|undefined>) => {
            setBuildingTrackHex(e.detail);
        };
        document.addEventListener('buildingTrackHex', handler);
        return () => document.removeEventListener('buildingTrackHex', handler);
    }, []);

    useEffect(() => {
        const handler = (e:CustomEventInit<MoveGoodsStep>) => {
            setPendingMoveGoods(e.detail);
        };
        document.addEventListener('pendingMoveGoods', handler);
        return () => document.removeEventListener('pendingMoveGoods', handler);
    }, []);

    let renderer = new RenderMapBuilder(game, map);

    let hexesWithTrack: { [ hexId: string]: boolean } = {};
    if (game.gameState) {
        // Render links
        if (game.gameState.links) {
            for (let link of game.gameState.links) {
                let hex = link.sourceHex;
                for (let i = 1; i < link.steps.length; i++) {
                    hex = applyDirection(hex, link.steps[i - 1]);
                    hexesWithTrack[hex.x + "," + hex.y] = true;
                }
            }
        }
    }

    for (let y = 0; y < map.getHeight(); y++) {
        for (let x = 0; x < map.getWidth(); x++) {
            renderer.renderHex({x: x, y: y});
        }
    }

    for (let interurbanLink of map.getInterurbanLinks()) {
        let owner: string|undefined;
        if (game.gameState && game.gameState.links) {
            for (let playerLink of game.gameState.links) {
                if (playerLink.sourceHex.x === interurbanLink.hex.x && playerLink.sourceHex.y === interurbanLink.hex.y
                    && playerLink.steps[0] === interurbanLink.direction) {
                    owner = playerLink.owner;
                    break;
                }
            }
        }
        if (pendingBuildAction && pendingBuildAction.interurbanLinkPlacements) {
            for (let playerLink of pendingBuildAction.interurbanLinkPlacements) {
                let other = applyDirection(playerLink.hex, playerLink.track);
                if ((playerLink.hex.x === interurbanLink.hex.x && playerLink.hex.y === interurbanLink.hex.y
                        && playerLink.track === interurbanLink.direction)
                    || (other.x === interurbanLink.hex.x && other.y === interurbanLink.hex.y
                        && oppositeDirection(playerLink.track) === interurbanLink.direction)) {
                    owner = game.activePlayer;
                    break;
                }
            }
        }
        if (owner) {
            renderer.renderOccupiedInterurbanLink(interurbanLink.hex, interurbanLink.direction, game.gameState?.playerColor[owner] as PlayerColor);
        } else {
            renderer.renderEmptyInterurbanLink(interurbanLink.hex, interurbanLink.direction, interurbanLink.cost);
        }
    }

    if (game.gameState) {
        let hexesWithTrack: { [ hexId: string]: boolean } = {};

        // Render links
        if (game.gameState.links) {
            for (let link of game.gameState.links) {
                let hex = link.sourceHex;
                let cityProperties = getCityProperties(game, map, hex);
                if (!cityProperties) {
                    if (map.getHexType(hex) !== HexType.TOWN) {
                        throw new Error("link started at non-city and non-town");
                    }
                    renderer.renderTownTrack(hex, link.steps[0], link.owner);
                }

                for (let i = 1; i < link.steps.length; i++) {
                    hex = applyDirection(hex, link.steps[i-1]);
                    let cityProperties = getCityProperties(game, map, hex);
                    if (!cityProperties) {
                        let left = oppositeDirection(link.steps[i-1]);
                        let right = link.steps[i];

                        // If this track is being redirected by a pending action, change "right" to match
                        if (!link.complete && pendingBuildAction && pendingBuildAction.trackRedirects) {
                            for (let pendingRedirect of pendingBuildAction.trackRedirects) {
                                if (pendingRedirect.hex.x === hex.x && pendingRedirect.hex.y === hex.y) {
                                    right = pendingRedirect.track;
                                }
                            }
                        }

                        if (map.getHexType(hex) === HexType.TOWN) {
                            renderer.renderTownTrack(hex, left, link.owner);
                        } else {
                            hexesWithTrack[hex.x + "," + hex.y] = true;
                            renderer.renderTrack(hex, left, right, link.owner);
                        }
                    }
                }

                // Render the last step in a completed link to a town
                hex = applyDirection(hex, link.steps[link.steps.length-1]);
                if (link.complete && map.getHexType(hex) === HexType.TOWN) {
                    let cityProperties = getCityProperties(game, map, hex);
                    if (!cityProperties) {
                        renderer.renderTownTrack(hex, oppositeDirection(link.steps[link.steps.length - 1]), link.owner);
                    }
                }
            }
        }

        // Render pending build action
        if (pendingBuildAction) {
            if (pendingBuildAction.urbanization) {
                renderer.renderCityHex(pendingBuildAction.urbanization.hex, urbCityProperties(pendingBuildAction.urbanization.city));
            }
            for (let townPlacement of pendingBuildAction.townPlacements) {
                renderer.renderTownTrack(townPlacement.hex, townPlacement.track, game.activePlayer);
            }
            for (let trackPlacement of pendingBuildAction.trackPlacements) {
                hexesWithTrack[trackPlacement.hex.x + "," + trackPlacement.hex.y] = true;
                renderer.renderTrack(trackPlacement.hex, trackPlacement.track[0], trackPlacement.track[1], game.activePlayer);
            }
        }

        for (let y = 0; y < map.getHeight(); y++) {
            for (let x = 0; x < map.getWidth(); x++) {
                let hex: Coordinate = {x: x, y: y};
                let specialCost = map.getSpecialTrackPricing(hex);
                if (specialCost !== undefined && !hexesWithTrack[hex.x + "," + hex.y]) {
                    renderer.renderSpecialCost(hex, specialCost);
                }
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
            renderer.renderActiveCube(pendingMoveGoods.currentCubePosition, pendingMoveGoods.selectedColor);
            if (pendingMoveGoods.nextStepOptions) {
                for (let option of pendingMoveGoods.nextStepOptions) {
                    let hex = applyDirection(pendingMoveGoods.currentCubePosition, option.direction);
                    renderer.renderArrow(hex, option.direction, option.owner);
                }
            }
        }

        if (buildingTrackHex) {
            for (let direction of ALL_DIRECTIONS) {
                let stepHex = applyDirection(buildingTrackHex, direction);
                let hexType = map.getHexType(stepHex);
                if (stepHex.x >= 0 && stepHex.y >= 0
                        && stepHex.x < map.getWidth() && stepHex.y < map.getHeight()
                        && hexType !== HexType.OFFBOARD
                        && hexType !== HexType.WATER) {
                    renderer.renderArrow(stepHex, direction, undefined);
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
