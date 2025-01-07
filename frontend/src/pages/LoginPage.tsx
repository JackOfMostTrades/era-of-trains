import UserSessionContext from "../UserSessionContext.tsx";
import {Container, Loader} from "semantic-ui-react";
import {useContext, useEffect} from "react";
import {Login} from "../api/api.ts";
import {useNavigate, useParams} from "react-router";
import ErrorContext from "../ErrorContext.tsx";

function LoginPage() {
    let navigate = useNavigate();
    let {reload} = useContext(UserSessionContext);
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
                        let res = await Login({provider: provider, accessToken: accessToken});
                        if (res.registrationRequired) {
                            let newParams = new URLSearchParams();
                            newParams.set("provider", provider);
                            newParams.set("access_token", accessToken);

                            window.location.href = '/register#' + newParams.toString();
                        } else {
                            window.location.hash = '';
                            await reload();
                            navigate('/');
                        }
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

export default LoginPage
