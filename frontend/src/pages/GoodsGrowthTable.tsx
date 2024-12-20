import {Header, Segment, Table, TableBody, TableCell, TableRow} from "semantic-ui-react";
import {Color, ViewGameResponse} from "../api/api.ts";
import './GoodsGrowthTable.css';
import {ReactNode} from "react";
import {colorToHtml} from "../actions/renderer/HexRenderer.tsx";

function emitEvent(x: number, y: number) {
    let event = new CustomEvent('goodsGrowthClickEvent', {
        detail: {
            x: x,
            y: y,
        }
    });
    document.dispatchEvent(event);
}

function GoodsGrowthTable({ game }: {game: ViewGameResponse}) {
    if (!game.gameState) {
        return
    }
    
    let goodsGrowth = game.gameState.goodsGrowth;

    let tableRows: ReactNode[] = [];

    let topHeaderCells: ReactNode[] = [];
    for (let i = 0; i < 6; i++) {
        topHeaderCells.push(<TableCell key={i}><div className="goodsGrowthHeader lightCity">{i+1}</div></TableCell>);
    }
    for (let i = 0; i < 6; i++) {
        topHeaderCells.push(<TableCell key={i+6}><div className="goodsGrowthHeader darkCity">{i+1}</div></TableCell>);
    }
    tableRows.push(<TableRow key={0}>{topHeaderCells}</TableRow>);
    for (let i = 0; i < 3; i++) {
        let cells: ReactNode[] = [];
        for (let j = 0; j < 12; j++) {
            let cube: ReactNode;
            if (goodsGrowth[j][i] !== Color.NONE) {
                let color = colorToHtml(goodsGrowth[j][i]);
                cube = <div className="cube" style={{background: color}} />;
            }
            cells.push(<TableCell key={j}><div className="cubeSpot" onClick={() => emitEvent(j,i)}>{cube}</div></TableCell>);
        }
        tableRows.push(<TableRow key={i+1}>{cells}</TableRow>);
    }

    let bottomHeaderCells: ReactNode[] = [];
    for (let i = 0; i < 4; i++) {
        bottomHeaderCells.push(<TableCell key={i}><div className="goodsGrowthHeader lightCity">{String.fromCharCode(i+'A'.charCodeAt(0))}</div></TableCell>);
    }
    for (let i = 0; i < 4; i++) {
        bottomHeaderCells.push(<TableCell key={i+4}><div className="goodsGrowthHeader darkCity">{String.fromCharCode(i+'E'.charCodeAt(0))}</div></TableCell>);
    }
    tableRows.push(<TableRow key={4}><TableCell/><TableCell/>{bottomHeaderCells}<TableCell/><TableCell/></TableRow>);

    for (let i = 0; i < 2; i++) {
        let cells: ReactNode[] = [];
        for (let j = 0; j < 8; j++) {
            let cube: ReactNode;
            if (goodsGrowth[j+12][i] !== Color.NONE) {
                let color = colorToHtml(goodsGrowth[j+12][i]);
                cube = <div className="cube" style={{background: color}} />;
            }
            cells.push(<TableCell key={j}><div className="cubeSpot" onClick={() => emitEvent(j+12,i)}>{cube}</div></TableCell>);
        }
        tableRows.push(<TableRow key={i+5}><TableCell/><TableCell/>{cells}<TableCell/><TableCell/></TableRow>);
    }
    
    return <Segment>
        <div style={{overflowX: "scroll"}}>
            <Header as='h2'>Goods Growth</Header>
            <Table celled unstackable>
                <TableBody>
                    {tableRows}
                </TableBody>
            </Table>
        </div>
    </Segment>
}

export default GoodsGrowthTable
