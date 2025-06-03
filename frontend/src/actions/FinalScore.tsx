import {Coordinate, GameState, User, ViewGameResponse} from "../api/api.ts";
import {GameMap, HexType, maps} from "../maps";
import {applyDirection, applyTeleport} from "../util.ts";
import {Header, Table, TableBody, TableCell, TableHeader, TableHeaderCell, TableRow} from "semantic-ui-react";

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

enum ScoreRow {
    INCOME_VPS,
    SHARE_VPS,
    TRACK_VPS,
    TOTAL_VPS,
}

function FinalScore({game}: {game: ViewGameResponse}) {
    if (!game.gameState) {
        return;
    }

    let map = maps[game.mapName];

    let scores: Map<ScoreRow, Map<string, number>> = new Map();
    for (const row of [ScoreRow.INCOME_VPS, ScoreRow.SHARE_VPS, ScoreRow.TRACK_VPS, ScoreRow.TOTAL_VPS]) {
        scores.set(row, new Map());
    }

    //let scores: { [playerId: string]: number} = {}
    for (let player of game.joinedUsers) {
        let income = game.gameState.playerIncome[player.id];
        if (income < 0) {
            scores.get(ScoreRow.TOTAL_VPS)!.set(player.id, -1);
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

        scores.get(ScoreRow.INCOME_VPS)!.set(player.id, income*3);
        scores.get(ScoreRow.SHARE_VPS)!.set(player.id, shares*-3);
        scores.get(ScoreRow.TRACK_VPS)!.set(player.id, trackCount);
        scores.get(ScoreRow.TOTAL_VPS)!.set(player.id, income*3 - shares*3 + trackCount);
    }

    let playerIds = [];
    let playersById: { [playerId: string]: User} = {};
    for (let player of game.joinedUsers) {
        playerIds.push(player.id);
        playersById[player.id] = player;
    }
    playerIds.sort((a, b) => {
        return scores.get(ScoreRow.TOTAL_VPS)!.get(b)! - scores.get(ScoreRow.TOTAL_VPS)!.get(a)!;
    });

    return <>
        <Header as='h2'>Final Scores</Header>
        <div style={{overflowX: "auto"}}>
            <Table celled compact unstackable>
                <TableHeader>
                    <TableRow>
                        <TableHeaderCell/>
                        {playerIds.map(playerId => <TableHeaderCell key={playerId}>{playersById[playerId].nickname}</TableHeaderCell>)}
                    </TableRow>
                </TableHeader>
                <TableBody>
                    <TableRow>
                        <TableCell>Income VPs</TableCell>
                        {playerIds.map(playerId => <TableCell key={playerId}>{scores.get(ScoreRow.TOTAL_VPS)!.get(playerId)! >= 0 ? scores.get(ScoreRow.INCOME_VPS)!.get(playerId) : ""}</TableCell>)}
                    </TableRow>
                    <TableRow>
                        <TableCell>Shares VPs</TableCell>
                        {playerIds.map(playerId => <TableCell key={playerId}>{scores.get(ScoreRow.TOTAL_VPS)!.get(playerId)! >= 0 ? scores.get(ScoreRow.SHARE_VPS)!.get(playerId) : ""}</TableCell>)}
                    </TableRow>
                    <TableRow>
                        <TableCell>Track VPs</TableCell>
                        {playerIds.map(playerId => <TableCell key={playerId}>{scores.get(ScoreRow.TOTAL_VPS)!.get(playerId)! >= 0 ? scores.get(ScoreRow.TRACK_VPS)!.get(playerId) : ""}</TableCell>)}
                    </TableRow>
                    <TableRow warning>
                        <TableCell>Total VPs</TableCell>
                        {playerIds.map(playerId => <TableCell key={playerId}>{scores.get(ScoreRow.TOTAL_VPS)!.get(playerId)! >= 0 ? scores.get(ScoreRow.TOTAL_VPS)!.get(playerId) : <span style={{fontStyle: "italic"}}>bankrupt</span>}</TableCell>)}
                    </TableRow>
                </TableBody>
            </Table>
        </div>
    </>
}

export default FinalScore
