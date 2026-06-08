# Sandboxing Infra Frontend

Deployable Next.js UI for sending JavaScript to a sandbox execution endpoint and viewing `out` and `err`.

Each run is treated as a standalone request. The backend can start a VM, execute the submitted code, return output, and stop the VM.

## Run locally

```sh
npm install
npm run dev
```

The app defaults to `http://localhost:3000/api/execute/js`. You can change the runner URL in the page, or set:

```sh
NEXT_PUBLIC_RUNNER_BASE_URL=https://runner.example.com
NEXT_PUBLIC_EXECUTE_ENDPOINT=/api/execute/js
```

For local development, copy `.env.example` to `.env.local` and adjust the values. The URL is not shown in the UI.

## Deploy

Deploy `apps/frontend` as the project root on Vercel, Netlify, or any host that supports Next.js.

The VM endpoint must accept browser requests and respond to:

```http
POST /api/execute/js
Content-Type: application/json

{ "code": "console.log('hello')" }
```

The UI supports either the current `{ "data": "..." }` response or separate `stdout` / `stderr` fields.
