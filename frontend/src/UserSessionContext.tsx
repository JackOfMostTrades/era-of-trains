import {createContext, ReactNode, useEffect, useState} from "react";
import {GetMyGames, Logout, WhoAmI, WhoAmIResponse} from "./api/api.ts";

interface UserSession {
    userInfo?: WhoAmIResponse
    waitingForMeCount: number
    loading?: boolean
    reload: () => Promise<void>
}

const UserSessionContext = createContext<UserSession>({waitingForMeCount: 0, reload: () => Promise.resolve()});

export function googleOauthSignin() {
    // Google's OAuth 2.0 endpoint for requesting an access token
    let oauth2Endpoint = 'https://accounts.google.com/o/oauth2/v2/auth';

    // Create <form> element to submit parameters to OAuth 2.0 endpoint.
    let form = document.createElement('form');
    form.setAttribute('method', 'GET'); // Send as a GET request.
    form.setAttribute('action', oauth2Endpoint);

    // eot-prod
    let clientId = '199571266655-hpirsmjge72vbahq532d8a205qilsmrl.apps.googleusercontent.com';
    if (window.location.hostname === 'localhost') {
        clientId = '199571266655-i6p30a84n6bodq3bj3nn1j5bmdlc714i.apps.googleusercontent.com';
    }

    // Parameters to pass to OAuth 2.0 endpoint.
    let params: {[key: string]: string} = {
        'client_id': clientId,
        'redirect_uri': window.location.origin + '/login/google',
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

export function discordOauthSignin(redirectPath: string) {
    let oauth2Endpoint = 'https://discord.com/oauth2/authorize';

    // Create <form> element to submit parameters to OAuth 2.0 endpoint.
    let form = document.createElement('form');
    form.setAttribute('method', 'GET'); // Send as a GET request.
    form.setAttribute('action', oauth2Endpoint);

    // eot-prod
    let clientId = '1326250803341692979';
    if (window.location.hostname === 'localhost') {
        clientId = '1326251048695894026';
    }

    // Parameters to pass to OAuth 2.0 endpoint.
    let params: {[key: string]: string} = {
        'client_id': clientId,
        'redirect_uri': window.location.origin + redirectPath,
        'response_type': 'token',
        'scope': 'identify email'};

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
    const reload: () => Promise<void> = async () => {
        try {
            let userInfoRes = await WhoAmI({});
            let myGamesRes = await GetMyGames({});
            let waitingForMeCount = 0;
            if (myGamesRes.games) {
                for (let game of myGamesRes.games) {
                    if (!game.finished && game.activePlayer === userInfoRes.user.id) {
                        waitingForMeCount += 1;
                    }
                }
            }
            setUserSession({loading: false, userInfo: userInfoRes, waitingForMeCount: waitingForMeCount, reload: reload});
        } catch (e) {
            setUserSession({loading: false, waitingForMeCount: 0, reload: reload});
        }
    }

    let [userSession, setUserSession] = useState<UserSession>({loading: true, waitingForMeCount: 0, reload: reload});

    useEffect(() => {
        reload();
    }, []);

    return <UserSessionContext.Provider value={userSession}>
        {children}
    </UserSessionContext.Provider>
}

export default UserSessionContext;
