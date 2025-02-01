import {useContext, useState} from "react";
import {Coordinate, ViewGameResponse} from "../api/api.ts";
import {
    Accordion,
    AccordionContent,
    AccordionTitle,
    Header,
    Icon,
    Segment,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from "semantic-ui-react";
import {MapTileState, MapTrackTile, TrackTile, TrackTiles} from "../game/map_state.ts";
import {GameMap, HexType} from "../maps";
import {HexRenderer} from "../actions/renderer/HexRenderer.tsx";
import UserSessionContext from "../UserSessionContext.tsx";

function TrackTileDisplay({ tile }: {tile: TrackTile}) {
    let userSession = useContext(UserSessionContext);
    let renderer = new HexRenderer(false, false, userSession);
    renderer.renderHex({x: 0, y: 0}, HexType.PLAINS);
    for (let route of new MapTrackTile(tile, 0).getRoutes()) {
        renderer.renderTrack({x: 0, y: 0}, route[0], route[1], undefined);
    }

    return renderer.render()
}

function PartsCountComponent({ map, game }: {map: GameMap, game: ViewGameResponse}) {
    let [open, setOpen] = useState<boolean>(false);

    if (!game.gameState) {
        return null;
    }

    let townMarkerCount = 8
    let tileCount = new Map<TrackTile, number>([
        [TrackTile.STRAIGHT_TRACK_TILE, 48],
        [TrackTile.SHARP_CURVE_TRACK_TILE, 7],
        [TrackTile.GENTLE_CURVE_TRACK_TILE, 55],

        [TrackTile.BOW_AND_ARROW_TRACK_TILE, 4],
        [TrackTile.TWO_GENTLE_TRACK_TILE, 3],
        [TrackTile.TWO_STRAIGHT_TRACK_TILE, 4],

        [TrackTile.BASEBALL_TRACK_TILE, 1],
        [TrackTile.LEFT_GENTLE_AND_SHARP_TRACK_TILE, 1],
        [TrackTile.RIGHT_GENTLE_AND_SHARP_TRACK_TILE, 1],
        [TrackTile.STRAIGHT_AND_SHARP_TRACK_TILE, 1],
    ]);

    let mapTileState = new MapTileState(map, game.gameState);
    for (let y = 0; y < map.getHeight(); y++) {
        for (let x = 0; x < map.getWidth(); x++) {
            let hex: Coordinate = {x: x, y: y};
            let tileState = mapTileState.getTileState(hex);
            if (tileState.isTown && !tileState.isCity && (tileState.routes.length == 2 || tileState.routes.length == 4)) {
                townMarkerCount -= 1;
            }

            let trackTile = MapTrackTile.fromRoutes(tileState.routes.map(r => [r.left, r.right]));
            if (!trackTile) {
                console.error("unable to identify track on tile (" + x + "," + y + ")");
            } else {
                let priorCount = tileCount.get(trackTile.getTile()) || 0;
                tileCount.set(trackTile.getTile(), priorCount - 1);
            }
        }
    }

    return <Segment>
        <Header as="h2">Components</Header>
        <Accordion>
            <AccordionTitle active={open} onClick={() => setOpen(!open)}>
                <Icon name='dropdown' />
                Components Count
            </AccordionTitle>
            <AccordionContent active={open} >
                <Table compact basic collapsing>
                    <TableBody>
                        {TrackTiles.map((tile,idx) => {
                            return <TableRow key={idx}>
                                <TableCell>
                                    <div style={{width: "4em"}}><TrackTileDisplay tile={tile} /></div>
                                </TableCell>
                                <TableCell>
                                    {tileCount.get(tile)}
                                </TableCell>
                            </TableRow>
                        })}
                    </TableBody>
                </Table>
                <p>Town markers: {townMarkerCount}</p>
            </AccordionContent>
        </Accordion>
    </Segment>
}

export default PartsCountComponent
