# Sandboxing Infra Frontend

Deployable Next.js UI for sending JavaScript to a sandbox execution endpoint and viewing `out` and `err`.

Each run is treated as a standalone request. The backend can start a VM, execute the submitted code, return output, and stop the VM.

## Run locally

```sh
npm install
npm run dev
```

The backend base URL is loaded from environment. Set:

```sh
NEXT_PUBLIC_RUNNER_BASE_URL=http://localhost:8000
NEXT_PUBLIC_EXECUTE_ENDPOINT=/api/execute
```

For local development, copy `.env.example` to `.env.local` and adjust the values. If `NEXT_PUBLIC_RUNNER_BASE_URL` is missing, the UI will show a configuration error instead of calling a fallback host.

## Deploy

Deploy `apps/frontend` as the project root on Vercel, Netlify, or any host that supports Next.js.

The VM backend endpoint must accept browser requests and respond to:

```http
POST /api/execute
Content-Type: application/json

{ "code": "console.log('hello')" }
```

The UI supports either the current `{ "data": "..." }` response or separate `stdout` / `stderr` fields.
