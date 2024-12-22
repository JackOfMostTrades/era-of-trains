import {BuildAction, Color, Coordinate, Direction, PlayerColor, ViewGameResponse} from "../api/api.ts";
import {ReactNode, useEffect, useState} from "react";
import maps, {BasicMap, HexType} from "../map.ts";
import {CityState, HexRenderer, urbCityState} from "../actions/renderer/HexRenderer.tsx";
import {applyDirection, oppositeDirection} from "../util.ts";
import {Step as MoveGoodsStep} from "../actions/MoveGoodsActionSelector.tsx";

function getCityState(game: ViewGameResponse, map: BasicMap, coordinate: Coordinate): CityState|undefined {
    if (game.gameState && game.gameState.urbanizations) {
        for (let urb of game.gameState.urbanizations) {
            if (urb.hex.x === coordinate.x && urb.hex.y === coordinate.y) {
                return urbCityState(urb.city);
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

class RenderMapBuilder {
    private game: ViewGameResponse;
    private map: BasicMap;
    private hexRenderer: HexRenderer;

    constructor(game: ViewGameResponse, map: BasicMap) {
        this.game = game;
        this.map = map;
        this.hexRenderer = new HexRenderer(true);
    }

    public renderCityHex(hex: Coordinate, cityState: CityState) {
        this.hexRenderer.renderCityHex(hex, cityState);
    }

    public renderHex(hex: Coordinate) {
        let cityState: CityState|undefined = getCityState(this.game, this.map, hex);
        if (cityState) {
            this.hexRenderer.renderCityHex(hex, cityState);
        } else {
            let hexType = this.map.hexes[hex.y][hex.x];
            this.hexRenderer.renderHex(hex, hexType);
        }
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

function ViewMapComponent({game}: {game: ViewGameResponse}) {
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

    let map = maps[game.mapName];
    let renderer = new RenderMapBuilder(game, map);

    for (let y = 0; y < map.height; y++) {
        for (let x = 0; x < map.width; x++) {
            renderer.renderHex({x: x, y: y});
        }
    }

    if (game.gameState) {
        // Render links
        if (game.gameState.links) {
            for (let link of game.gameState.links) {
                let hex = link.sourceHex;
                let cityState = getCityState(game, map, hex);
                if (!cityState) {
                    if (map.hexes[hex.y][hex.x] !== HexType.TOWN) {
                        throw new Error("link started at non-city and non-town");
                    }
                    renderer.renderTownTrack(hex, link.steps[0], link.owner);
                }

                for (let i = 1; i < link.steps.length; i++) {
                    hex = applyDirection(hex, link.steps[i-1]);
                    let cityState = getCityState(game, map, hex);
                    if (!cityState) {
                        let left = oppositeDirection(link.steps[i-1]);
                        let right = link.steps[i];

                        if (map.hexes[hex.y][hex.x] === HexType.TOWN) {
                            renderer.renderTownTrack(hex, left, link.owner);
                        } else {
                            renderer.renderTrack(hex, left, right, link.owner);
                        }
                    }
                }

                // Render the last step in a completed link to a city
                hex = applyDirection(hex, link.steps[link.steps.length-1]);
                if (link.complete && map.hexes[hex.y][hex.x] === HexType.TOWN) {
                    let cityState = getCityState(game, map, hex);
                    if (!cityState) {
                        renderer.renderTownTrack(hex, oppositeDirection(link.steps[link.steps.length - 1]), link.owner);
                    }
                }
            }
        }

        // Render pending build action
        if (pendingBuildAction) {
            if (pendingBuildAction.urbanization) {
                renderer.renderCityHex(pendingBuildAction.urbanization.hex, urbCityState(pendingBuildAction.urbanization.city));
            }
            for (let townPlacement of pendingBuildAction.townPlacements) {
                for (let track of townPlacement.tracks) {
                    renderer.renderTownTrack(townPlacement.hex, track, game.activePlayer);
                }
            }
            for (let trackPlacement of pendingBuildAction.trackPlacements) {
                for (let track of trackPlacement.tracks) {
                    renderer.renderTrack(trackPlacement.hex, track[0], track[1], game.activePlayer);
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
    }

    return <>
        <div style={{marginTop: '1em'}} />
        {renderer.render()}
    </>
}

export default ViewMapComponent;
