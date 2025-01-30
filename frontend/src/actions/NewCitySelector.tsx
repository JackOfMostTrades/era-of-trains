import {ReactNode, useContext} from "react";
import {HexRenderer, urbCityProperties} from "./renderer/HexRenderer.tsx";
import {Container} from "semantic-ui-react";
import UserSessionContext from "../UserSessionContext.tsx";

interface NewCitySelectorProps {
    selected: number
    alreadyUrbanized: number[]
    onChange: (selected: number) => void
}

export function NewCitySelector(props: NewCitySelectorProps) {
    let userSession = useContext(UserSessionContext);

    let columns: ReactNode[] = [];
    for (let newCityNum = 0; newCityNum < 8; newCityNum++) {
        let renderer = new HexRenderer(false, false, userSession);
        renderer.renderCityHex({x: 0, y: 0}, urbCityProperties(newCityNum));

        let classNames = "track-select";
        if (newCityNum === props.selected) {
            classNames += " selected";
        }
        let available = props.alreadyUrbanized.indexOf(newCityNum) === -1;
        let onClick: undefined | (() => void);
        if (available) {
            onClick = () => props.onChange(newCityNum);
        } else {
            classNames += " unavailable";
        }

        columns.push(<div className={classNames} onClick={onClick}>{renderer.render()}</div>)
    }
    return <Container>
        {columns}
    </Container>
}
