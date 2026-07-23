# Vault API

![Status of tests on this REPO](https://github.com/mrbaker1917/vault_api/actions/workflows/ci.yml/badge.svg)

Zero-knowledge password vault backend in Go. The server authenticates users and stores encrypted blobs plus metadata; clients derive keys and encrypt vault contents locally. Also, a frontend using TypeScript and Tailwind.

**Status:** MVP complete — auth, vault CRUD, MFA, recovery, sharing, audit, metrics, CI, and docs are implemented.

## Motivation: 
I wanted to create a password manager to understand how it would be constructed, considering all the possible security problems.

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+ (or Docker)
- For integration tests: Docker (testcontainers)

## Contributing

```bash
# Apply migrations to your Postgres instance (goose), then:
export DATABASE_URL="postgres://vault:vault@localhost:5433/vault_api?sslmode=disable"
export JWT_SECRET="dev-secret-change-me-for-local-only!!"
export APP_ENV=development
export PORT=8081

go run ./cmd/server
```

Health checks:

- `GET /health` — liveness
- `GET /ready` — readiness (PostgreSQL ping)
- `GET /metrics` — Prometheus metrics

### Docker Compose

```bash
cd docker
docker compose up --build
```

The API listens on [http://localhost:8081](http://localhost:8081). Apply migrations to Postgres before first use (not automated on startup yet).

### Web frontend

```bash
cd web
cp .env.example .env
npm install
npm run dev
```

The React app runs at [http://localhost:5173](http://localhost:5173). See [web/README.md](web/README.md).

## Usage:

### Implemented

| Area | Details |
|------|---------|
| **Auth** | Signup, login, logout, refresh, **change password**, JWT + DB-backed sessions |
| **Sessions** | List and revoke device sessions |
| **MFA** | TOTP enable / verify / disable |
| **Recovery** | One-time recovery codes (requires MFA) |
| **Vault** | CRUD, pagination, filtering, optimistic locking, soft delete / restore |
| **Sharing** | Share items by email with client-wrapped keys (`read` / `write`) |
| **Audit** | Append-only log of sensitive operations |
| **Security** | Argon2id passwords, rate-limited auth, encrypted blob validation |
| **Ops** | JSON logging, `/health`, `/ready`, Prometheus `/metrics`, Docker, GitHub Actions CI |
| **Web UI** | React app in `web/` — auth, encrypted vault CRUD, MFA, recovery, sessions, settings, audit log, trash restore |

### Planned / not yet implemented

- Vault key re-wrapping API (client-driven master password change)
- Automated purge of soft-deleted items after 30 days
- Breach-aware password checks (e.g. HIBP)
- Shared-user `write` permission enforcement on vault updates
- Redis-backed sessions / distributed rate limiting (Redis is in Compose but unused)
- Auto-migrations on app startup
- Demo client (CLI or web) showing client-side encryption

## Zero-knowledge model

```
Client                         Backend                    Database
  |                               |                            |
  |-- master password (local)     |                            |
  |-- encrypt vault item          |                            |
  |-- POST encrypted blob + meta ->|-- store ciphertext + meta ->|
  |<- GET encrypted blob + meta --|<-- fetch -------------------|
  |-- decrypt locally             |                            |
```

- **Server never sees** the master password or plaintext vault secrets.
- **Server stores** ciphertext, titles, folders, tags, and item types (metadata enables search without decryption).
- **Client responsibility** — key derivation, encryption, and share key wrapping. See [Architecture](docs/architecture.md).

Account authentication (Argon2id, MFA, sessions) is separate from vault encryption.

## API

Full contract: [`docs/openapi.yaml`](docs/openapi.yaml) (validated in CI).

| Group | Endpoints |
|-------|-----------|
| Auth | `POST /api/v1/auth/signup`, `/login`, `/logout`, `/refresh`, `/change-password` · `GET /api/v1/me` · `GET/DELETE /api/v1/auth/sessions` |
| MFA | `POST /api/v1/mfa/enable`, `/verify`, `/disable` |
| Recovery | `POST /api/v1/recovery/generate`, `/verify` |
| Vault | `POST/GET/PUT/DELETE /api/v1/vault/items` · `GET /api/v1/vault/items/deleted` · `POST .../restore` |
| Sharing | `POST .../share` · `DELETE .../share/{user_id}` · `GET /api/v1/vault/shared` |
| Audit | `GET /api/v1/audit/logs` |
| Ops | `GET /health`, `/ready`, `/metrics` |

Protected routes require `Authorization: Bearer <access_token>`.

## Project structure

```
vault_api/
├── cmd/server/           # Entry point
├── internal/
│   ├── api/              # Router, handlers, middleware
│   ├── service/          # Business logic
│   ├── repository/       # Postgres (pgx + sqlc)
│   ├── domain/           # Entity types
│   ├── crypto/           # Hashing, JWT, TOTP, blob validation
│   └── config/
├── migrations/           # Goose SQL migrations
├── sql/queries/          # sqlc query definitions
├── docker/               # Dockerfile + docker-compose.yml
├── docs/                 # Architecture, threat model, OpenAPI
└── .github/workflows/    # CI
```

## Tech stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.25 |
| HTTP | `net/http` (Go 1.22+ routing) |
| Database | PostgreSQL 17, pgx/v5, sqlc |
| Migrations | Goose |
| Password hashing | Argon2id |
| Auth | JWT (HS256) + refresh tokens (SHA-256 hashed in DB) |
| MFA | TOTP (pquerna/otp) |
| Metrics | Prometheus client |
| Logging | `log/slog` (JSON) |
| Tests | testcontainers-go (integration), race detector in CI |

## Development

```bash
go mod verify
golangci-lint run ./...
go test -race ./...
go test -tags=integration -race ./internal/api/...
go build -o bin/server ./cmd/server
```

CI (`.github/workflows/ci.yml`) runs lint, unit tests, integration tests, OpenAPI validation, and build on every push/PR.

## Documentation

- [Architecture](docs/architecture.md) — system design, data model, auth and vault flow
- [Threat model](docs/threat_model.md) — assets, trust boundaries, threats, mitigations
- [OpenAPI spec](docs/openapi.yaml) — HTTP API contract

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP listen port |
| `APP_ENV` | `production` | Set to `development` / `dev` / `local` / `test` to allow weak JWT secrets locally |
| `DATABASE_URL` | local Postgres DSN | PostgreSQL connection string |
| `JWT_SECRET` | `change-me` | HS256 signing key; **required ≥32 chars** and not a known weak value unless `APP_ENV` is development |
| `REDIS_URL` | `redis://localhost:6379` | Reserved (not used yet) |
| `CORS_ALLOWED_ORIGINS` | — | Comma-separated allowed origins |

## References

- [Bitwarden Security White Paper](https://bitwarden.com/help/bitwarden-security-white-paper/) — zero-knowledge design patterns
- [Go project layout](https://github.com/golang-standards/project-layout)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [RFC 6238 — TOTP](https://datatracker.ietf.org/doc/html/rfc6238)
