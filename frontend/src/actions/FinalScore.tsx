import {Coordinate, GameState, User, ViewGameResponse} from "../api/api.ts";
import {GameMap, HexType, maps} from "../maps";
import {applyDirection, applyTeleport} from "../util.ts";
import {Header, List, ListItem} from "semantic-ui-react";

function isCity(gameState: GameState, map: GameMap, hex: Coordinate): boolean {
    if (map.getHexType(hex) === HexType.CITY) {
        return true;
    }
    if (gameState.urbanizations) {
        for (let urb of gameState.urbanizations) {
            if (urb.hex.x === hex.x && urb.hex.y === hex.y) {
                return true;
            }
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
        let income = game.gameState.playerIncome[player.id];
        if (income < 0) {
            scores[player.id] = -1;
            continue;
        }

        let shares = game.gameState.playerShares[player.id];

        let trackCount = 0;
        for (let link of game.gameState.links) {
            if (!link.complete || link.owner !== player.id) {
                continue;
            }

            let hex = link.sourceHex;
            // Count steps that aren't part of a city (always adding one for teleport links)
            for (let i = 0; i < link.steps.length; i++) {
                if (!isCity(game.gameState, map, hex)) {
                    trackCount += 1;
                }
                let teleportEdge = applyTeleport(map, game.gameState, undefined, hex, link.steps[i]);
                let nextHex: Coordinate;
                if (teleportEdge !== undefined) {
                    trackCount += 1;
                    nextHex = teleportEdge.hex;
                } else {
                    nextHex = applyDirection(hex, link.steps[i]);
                }
                hex = nextHex;
            }
            if (!isCity(game.gameState, map, hex)) {
                trackCount += 1;
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
                let score = scores[playerId];
                return <ListItem>{playersById[playerId].nickname}: {score < 0 ? <span style={{fontStyle: "italic"}}>bankrupt</span> : <span>{score}</span>}</ListItem>
            })}
        </List>
    </>
}

export default FinalScore
