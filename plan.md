# Vault API Implementation Plan

## 1. Bootstrap the Project Skeleton
1. Initialize Go module and package layout (`cmd`, `internal`, `migrations`, `docker`, `docs`).
2. Add a minimal HTTP server with `GET /health`.
3. Implement config loading from environment variables (`PORT`, `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`).
4. Add Docker Compose for app, PostgreSQL, and Redis.

Why first: This establishes a runnable baseline so every next feature lands on working infrastructure.

## 2. Implement Database and Migrations
1. Create migrations for `users`, `sessions`, `vault_items`, `recovery_codes`, and `audit_log`.
2. Add a migration runner command.
3. Implement DB connection layer with startup ping and retry.
4. Define repository interfaces and implement `users` and `sessions` repositories first.

Why first: Authentication and vault features depend on persistent state and stable schema contracts.

## 3. Build Authentication First (Vertical Slice)
1. Add password hashing utility (Argon2id preferred, bcrypt acceptable).
2. Add token generation and validation (JWT or opaque token with hashed DB storage).
3. Implement `POST /api/v1/auth/signup` end-to-end.
4. Implement `POST /api/v1/auth/login` end-to-end.
5. Add auth middleware for protected routes.

Why first: Most other capabilities require identity and authorization.

## 4. Add Core Middleware and Request Safety
1. Structured JSON request logging.
2. Panic recovery middleware.
3. CORS policy.
4. Rate limiting on auth endpoints.

Why first: Prevents insecure or unobservable endpoint growth that needs later rework.

## 5. Build Vault CRUD (Core Product Value)
1. `POST /api/v1/vault/items` (store encrypted blob and metadata only).
2. `GET /api/v1/vault/items` (pagination, filtering).
3. `GET /api/v1/vault/items/{id}` (ownership checks).
4. `PUT /api/v1/vault/items/{id}` (version bump, optimistic update handling).
5. `DELETE /api/v1/vault/items/{id}` soft delete and `POST /api/v1/vault/items/{id}/restore` restore.

Why first: This proves the zero-knowledge backend boundary and delivers MVP functionality.

## 6. Add Session and Device Management
1. Persist device session metadata.
2. Implement `GET /api/v1/auth/sessions`.
3. Implement `DELETE /api/v1/auth/sessions/{id}`.
4. Implement `POST /api/v1/auth/logout` to revoke current session.

Why first: Session visibility and revocation are key security controls for password managers.

## 7. Implement Security Hardening Features
1. MFA setup and verification (TOTP).
2. Recovery code generation and one-time consumption.
3. Audit log service for all sensitive operations.
4. Encrypted blob structure validation.
5. Password strength and breach-aware checks.

Why first: These are the highest-value security differentiators for production readiness and interviews.

## 8. Implement Sharing Flows
1. Share vault item with another user (`encrypted_item_key`, permission).
2. Revoke sharing.
3. List items shared with current user.

Why first: Sharing introduces more complex authorization; safer after ownership model is stable.

## 9. Add Operations and Delivery Readiness
1. Prometheus metrics (`/metrics`).
2. Health and readiness checks.
3. CI pipeline (lint, unit tests, integration tests, migration check).
4. OpenAPI documentation updates.
5. Architecture and threat model documentation.

Why first: Observability, automation, and documentation complete production-grade delivery.

## First Week Execution Checklist
1. Day 1: Server bootstrap, `/health`, env config, Docker Compose.
2. Day 2: Migrations, DB connection, initial user repository.
3. Day 3: Signup/login, hashing, token issue and validation.
4. Day 4: Auth middleware, protected test route, baseline tests.
5. Day 5: Vault create/list/get with ownership enforcement.

## Definition of Done for MVP
1. Auth works with secure password hashing and session/token management.
2. Vault CRUD stores encrypted payloads only, never plaintext secrets.
3. Soft delete and restore workflows function correctly.
4. Audit logging records sensitive actions.
5. Basic observability (`/health`, structured logs, `/metrics`) is available.
6. CI validates linting, tests, and migrations.


psql "postgres://vault:vault@localhost:5433/vault_api?sslmode=disable"

goose -dir migrations postgres "postgres://vault:vault@localhost:5433/vault_api?sslmode=disable" up
goose -dir migrations postgres "postgres://vault:vault@localhost:5433/vault_api?sslmode=disable" down