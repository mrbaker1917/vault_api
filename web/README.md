# Vault Web

React frontend for the Vault API.

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

The API allows CORS from `http://localhost:5173` and `http://127.0.0.1:5173` by default.

## Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Vite dev server |
| `npm run build` | Production build |
| `npm run test` | Crypto unit tests |
| `npm run preview` | Preview production build |

## Phase 1 — Authentication

- Signup and login (including MFA prompt when required)
- JWT storage in `sessionStorage`
- Automatic access-token refresh
- Protected app shell

## Phase 2 — Crypto + vault

- **Master password** (separate from account password) — never sent to the server
- PBKDF2 key derivation + AES-256-GCM encryption in the browser
- Encrypted blob format: `[0x01 version][12-byte IV][ciphertext]` (matches API validation)
- Vault unlock / lock flow with local salt + verifier in `localStorage`
- Vault item list, create, edit, delete (login, note, card, identity types)
- Optimistic locking via `version` on update/delete

Metadata (`title`, `folder`, `tags`, `item_type`) is stored in plaintext on the server for search and listing. Only secret fields inside `encrypted_data` are protected client-side.

## Next (Phase 3)

- MFA setup UI, recovery codes, sessions, account password change
