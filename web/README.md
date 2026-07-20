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
- **Portable salt** derived from your user ID — same master password works on any browser/device
- Legacy per-browser salts are migrated automatically on unlock
- Encrypted blob format: `[0x01 version][12-byte IV][ciphertext]` (matches API validation)
- Vault unlock / lock flow with local salt + verifier in `localStorage`
- Vault item list, create, edit, delete (login, note, card, identity types)
- **Detail view** — read-only item screen with copy buttons, show/hide secrets, and open URL for logins
- Optimistic locking via `version` on update/delete

Metadata (`title`, `folder`, `tags`, `item_type`) is stored in plaintext on the server for search and listing. Only secret fields inside `encrypted_data` are protected client-side.

## Phase 3 — Security settings

- MFA enrollment (QR code + TOTP verify) and disable
- Recovery code generation (requires MFA) and recovery sign-in page
- Active sessions list with revoke
- Account password change (with TOTP when MFA enabled)

Settings are at `/settings` and do not require vault unlock.

## Phase 4 — Audit, trash, deploy

- **Audit log** at `/audit` — paginated activity from `GET /api/v1/audit/logs`
- **Trash** at `/trash` — list soft-deleted items and restore them (no vault unlock required)
- New API: `GET /api/v1/vault/items/deleted`

### Production build

```bash
npm run build
```

Serve the `dist/` folder with any static host. Set the API URL at build time:

```bash
VITE_API_URL=https://api.example.com npm run build
```

Ensure the API allows your frontend origin via `CORS_ALLOWED_ORIGINS`.
