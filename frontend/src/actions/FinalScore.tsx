import {Coordinate, GameState, User, ViewGameResponse} from "../api/api.ts";
import {GameMap, HexType, maps} from "../maps";
import {applyDirection} from "../util.ts";
import {Header, List, ListItem} from "semantic-ui-react";

function isCity(gameState: GameState, map: GameMap, hex: Coordinate): boolean {
    if (map.getHexType(hex) === HexType.CITY) {
        return true;
    }
    for (let urb of gameState.urbanizations) {
        if (urb.hex.x === hex.x && urb.hex.y === hex.y) {
            return true;
        }
    }

    return false;
}

function FinalScore({game}: {game: ViewGameResponse}) {
    if (!game.gameState) {
        return;
    }

    let map = maps[game.mapName];

    let scores: { [playerId: string]: number} = {}
    for (let player of game.joinedUsers) {
        let shares = game.gameState.playerShares[player.id];
        let income = game.gameState.playerIncome[player.id];
        let trackCount = 0;
        for (let link of game.gameState.links) {
            if (!link.complete || link.owner !== player.id) {
                continue;
            }

            let hex = link.sourceHex;
            // If it's an interurban link, add 1
            if (link.steps.length === 1
                && isCity(game.gameState, map, hex)
                && isCity(game.gameState, map, applyDirection(hex, link.steps[0]))) {
                trackCount += 1;
            } else {
                // Otherwise count steps that aren't part of a city
                for (let i = 0; i < link.steps.length; i++) {
                    if (!isCity(game.gameState, map, hex)) {
                        trackCount += 1;
                    }
                    hex = applyDirection(hex, link.steps[i]);
                }
                if (!isCity(game.gameState, map, hex)) {
                    trackCount += 1;
                }
            }
        }

        scores[player.id] = income*3 - shares*3 + trackCount;
    }

    let playerIds = [];
    let playersById: { [playerId: string]: User} = {};
    for (let player of game.joinedUsers) {
        playerIds.push(player.id);
        playersById[player.id] = player;
    }
    playerIds.sort((a, b) => {
        return scores[b] - scores[a];
    });

    return <>
        <Header as='h2'>Final Scores</Header>
        <List>
            {playerIds.map(playerId => {
                return <ListItem>{playersById[playerId].nickname}: {scores[playerId]}</ListItem>
            })}
        </List>
    </>
}

export default FinalScore
