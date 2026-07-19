# Vault Web

React frontend for the Vault API (Phase 1: authentication).

## Setup

```bash
cp .env.example .env
npm install
```

## Development

Start the API on port 8081, then:

```bash
npm run dev
```

Open [http://localhost:5173](http://localhost:5173).

The API allows CORS from `http://localhost:5173` by default. Override with
`CORS_ALLOWED_ORIGINS` on the server if needed.

## Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Vite dev server |
| `npm run build` | Production build |
| `npm run preview` | Preview production build |

## Phase 1 scope

- Signup and login (including MFA prompt when required)
- JWT storage in `sessionStorage`
- Automatic access-token refresh
- Protected dashboard shell

Vault encryption and item management arrive in Phase 2.
