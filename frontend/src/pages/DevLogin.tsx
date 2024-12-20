import {Container} from "semantic-ui-react";
import {useNavigate, useParams} from "react-router";
import {Login} from "../api/api.ts";
import {useContext, useEffect} from "react";
import ErrorContext from "../ErrorContext.tsx";

function DevLogin() {
    let params = useParams();
    let navigate = useNavigate();
    let {setError} = useContext(ErrorContext);

    let nickname = params.nickname;

    useEffect(() => {
        if (nickname) {
            Login({
                devNickname: nickname
            }).then(() => {
                return navigate('/');
            }).then(() => {
                return navigate(0)
            }).catch(err => {
                setError(err);
            })
        }
    }, [nickname]);

    return <Container text>
        <p>Signing in...</p>
    </Container>
}

export default DevLogin
