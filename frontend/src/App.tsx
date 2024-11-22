import {useEffect, useState} from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

/*
 * Create form to request access token from Google's OAuth 2.0 server.
 */
function oauthSignIn() {
    // Google's OAuth 2.0 endpoint for requesting an access token
    let oauth2Endpoint = 'https://accounts.google.com/o/oauth2/v2/auth';

    // Create <form> element to submit parameters to OAuth 2.0 endpoint.
    let form = document.createElement('form');
    form.setAttribute('method', 'GET'); // Send as a GET request.
    form.setAttribute('action', oauth2Endpoint);

    // Parameters to pass to OAuth 2.0 endpoint.
    let params: {[key: string]: string} = {
        'client_id': '902198735009-tqffhnfnc5smc1sp342fot322559ns4v.apps.googleusercontent.com',
        'redirect_uri': window.location.href,
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

function App() {
  const [count, setCount] = useState(0)
  const [username, setUsername] = useState();

  useEffect(() => {
    async function handleLogin() {
        let h = window.location.hash;
        if (h.charAt(0) != '#') {
            return;
        }
        let q = new URLSearchParams(h.substring(1));
        let accessToken = q.get('access_token');
        if (accessToken) {
            try {
                let res = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'same-origin',
                    body: JSON.stringify({"accessToken": accessToken})
                });
                if (!res.ok) {
                    throw new Error("Got non-ok status: " + res.status);
                }
                res = await fetch('/api/whoami', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'same-origin',
                    body: JSON.stringify({})
                });
                if (!res.ok) {
                    throw new Error("Got non-ok status: " + res.status);
                }
                let loginData = await res.json();
                history.pushState("", document.title, window.location.pathname + window.location.search);
                setUsername(loginData.user);
            } catch (e) {
                console.error("Unhandled error logging in: " + e);
            }
        }
    }

    handleLogin();
  }, []);

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <br/>
        <button onClick={() => oauthSignIn()}>Sign-In</button>
        {username ? <p>Loggin-in as: {username}</p> : null}
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  )
}

export default App
