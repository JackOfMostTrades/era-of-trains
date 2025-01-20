# Development

To do development, you must run both the frontend and backend.
To start the backend, go into the `backend` directory and run `go run .`.
To start the frontend, go into the `frontend` directory and run `npm run dev`.

Browse to `http://localhost:5173` to test.
You will need to authenticate using real google accounts.

## Bootstrap

To bootstrap the database, update backend/config-local.json to set `database.bootstrap` to true.
