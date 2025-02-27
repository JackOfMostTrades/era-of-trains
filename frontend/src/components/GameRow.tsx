import {GameSummary} from "../api/api.ts";
import {TableCell, TableRow} from "semantic-ui-react";
import {Link} from "react-router";
import {mapNameToDisplayName} from "../util.ts";

function GameRow({ game }: {game: GameSummary}) {
    let status: string;
    if (!game.started) {
        if (game.joinedUsers.length < game.minPlayers) {
            status = 'Waiting for players';
        } else if (game.joinedUsers.length < game.maxPlayers) {
            status = 'Open to more players, but ready to start';
        } else {
            status = 'Game full; waiting to start';
        }
    } else if (game.finished) {
        status = 'Finished';
    } else {
        status = 'In progress';
    }

    let playerCount: number | string;
    if (game.started) {
        playerCount = game.joinedUsers.length;
    } else {
        if (game.minPlayers === game.maxPlayers) {
            playerCount = game.minPlayers;
        } else {
            playerCount = game.minPlayers + " to " + game.maxPlayers
        }
    }

    let activePlayer: string|undefined;
    for (let user of game.joinedUsers) {
        if (user.id === game.activePlayer) {
            activePlayer = user.nickname;
            break;
        }
    }

    return <TableRow>
        <TableCell><Link to={`/games/${game.id}`}>{game.name}</Link></TableCell>
        <TableCell>{playerCount}</TableCell>
        <TableCell>{mapNameToDisplayName(game.mapName)}</TableCell>
        <TableCell>{status}</TableCell>
        <TableCell>{activePlayer}</TableCell>
    </TableRow>
}

export default GameRow;
