import {Icon, Message, MessageHeader} from "semantic-ui-react";
import {useContext} from "react";
import ErrorContext from "../ErrorContext.tsx";

function ErrorDisplay() {
    let {error, setError} = useContext(ErrorContext);
    if (!error) {
        return null;
    }
    return <Message negative>
        <div style={{cursor: "pointer", float: "right"}} onClick={() => setError(undefined)}><Icon name="close" /></div>
        <MessageHeader>Something went wrong!</MessageHeader>
        <p>{'' + error}</p>
    </Message>
}

export default ErrorDisplay;
