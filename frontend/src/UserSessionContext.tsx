import {createContext, useEffect, useState} from "react";

interface UserSession {
    userInfo?: {User: string}
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
    let res = await fetch('/api/logout', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({})
    });
    if (!res) {
        throw new Error("failed to logout");
    }
    window.location.reload();
}

export function UserSessionProvider({ children }) {
    let [userSession, setUserSession] = useState<UserSession>({loading: true});

    useEffect(() => {
        (async () => {
            if (window.location.hash) {
                let params = new URLSearchParams(window.location.hash.substring(1));
                let accessToken = params.get("access_token");
                if (accessToken) {
                    let res = await fetch('/api/login', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({"accessToken": accessToken})
                    });
                    if (!res) {
                        throw new Error("failed to login");
                    }
                    window.location.hash = '';
                }
            }

            let res = await fetch('/api/whoami', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({})
            });
            if (!res.ok) {
                setUserSession({loading: false});
                return;
            }
            let user = await res.json();
            setUserSession({loading: false, userInfo: user});
        })();
    }, []);

    return <UserSessionContext.Provider value={userSession}>
        {children}
    </UserSessionContext.Provider>
}

export default UserSessionContext;
