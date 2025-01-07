import {Container, Loader} from "semantic-ui-react";
import {useContext, useEffect} from "react";
import {LinkProfile} from "../api/api.ts";
import {useNavigate, useParams} from "react-router";
import ErrorContext from "../ErrorContext.tsx";

function LinkProfilePage() {
    let navigate = useNavigate();
    let {setError} = useContext(ErrorContext);
    let params = useParams();

    let provider = params.provider;
    useEffect(() => {
        (async () => {
            if (window.location.hash) {
                let params = new URLSearchParams(window.location.hash.substring(1));
                let accessToken = params.get("access_token");
                if (provider && accessToken) {
                    try {
                        await LinkProfile({provider: provider, accessToken: accessToken});
                        navigate('/profile');
                    } catch (e) {
                        setError(e);
                        return;
                    }
                }
            }
        })();
    }, []);

    return <Container text>
        <Loader active />
    </Container>
}

export default LinkProfilePage
