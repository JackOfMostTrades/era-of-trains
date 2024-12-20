import {createContext, ReactNode, useContext, useEffect, useState} from "react";
import {Login, Logout, WhoAmI, WhoAmIResponse} from "./api/api.ts";
import ErrorContext from "./ErrorContext.tsx";

interface UserSession {
    userInfo?: WhoAmIResponse
    loading?: boolean
}

const UserSessionContext = createContext<UserSession>({});

export function oauthSignIn() {
    // Google's OAuth 2.0 endpoint for requesting an access token
    let oauth2Endpoint = 'https://accounts.google.com/o/oauth2/v2/auth';

    // Create <form> element to submit parameters to OAuth 2.0 endpoint.
    let form = document.createElement('form');
    form.setAttribute('method', 'GET'); // Send as a GET request.
    form.setAttribute('action', oauth2Endpoint);

    // Parameters to pass to OAuth 2.0 endpoint.
    let params: {[key: string]: string} = {
        'client_id': '902198735009-tqffhnfnc5smc1sp342fot322559ns4v.apps.googleusercontent.com',
        'redirect_uri': window.location.origin + '/',
        'response_type': 'token',
        'scope': 'https://www.googleapis.com/auth/userinfo.email',
        'include_granted_scopes': 'true'};

    // Add form parameters as hidden input values.
    for (let p in params) {
        let input = document.createElement('input');
        input.setAttribute('type', 'hidden');
        input.setAttribute('name', p);
        input.setAttribute('value', params[p]);
        form.appendChild(input);
    }

    // Add form to page and submit it to open the OAuth 2.0 endpoint.
    document.body.appendChild(form);
    form.submit();
}

export async function logout() {
    await Logout({});
    window.location.reload();
}

export function UserSessionProvider({ children }: {children: ReactNode}) {
    let [userSession, setUserSession] = useState<UserSession>({loading: true});
    let {setError} = useContext(ErrorContext);

    useEffect(() => {
        (async () => {
            if (window.location.hash) {
                let params = new URLSearchParams(window.location.hash.substring(1));
                let accessToken = params.get("access_token");
                if (accessToken) {
                    try {
                        await Login({accessToken: accessToken});
                        window.location.hash = '';
                    } catch (e) {
                        setError(e);
                        return;
                    }
                }
            }

            try {
                let res = await WhoAmI({});
                setUserSession({loading: false, userInfo: res});
            } catch (e) {
                setUserSession({loading: false});
                return;
            }
        })();
    }, []);

    return <UserSessionContext.Provider value={userSession}>
        {children}
    </UserSessionContext.Provider>
}

export default UserSessionContext;
