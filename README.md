# vault_api

Vault API: Zero-Knowledge Password Manager Backend
Project Overview
Build a secure, production-grade password vault backend in Go that demonstrates cryptographic design, authentication, audit logging, and clean API architecture. The system uses zero-knowledge architecture where the server never sees plaintext secrets.
Target audience: Password management companies (e.g., 1Password, Bitwarden, LastPass)
Tech stack: Go, PostgreSQL, Redis, Docker
Timeline: 4-6 weeks for MVP

Core Features
Authentication & User Management
User signup with email/password
Secure login with JWT or session tokens
Password hashing with bcrypt or Argon2
Multi-factor authentication (TOTP)
Device session management
Account recovery with recovery codes
Vault Operations
Create, read, update, delete vault items
Client-side encryption with user master key
Server stores only encrypted blobs + metadata
Search vault items by metadata (tags, titles)
Organize items into folders/categories
Soft delete with 30-day recovery window
Security Features
Zero-knowledge architecture (server never sees plaintext)
Per-user encryption keys derived from master password
Key re-wrapping for password changes
Secure vault item sharing between users
Rate limiting on auth endpoints
Breach-aware password strength checks
Audit log for all sensitive operations
Operational Features
Health check endpoints
Structured logging (JSON)
Prometheus metrics
Request tracing
Database migrations
Docker Compose deployment
CI/CD pipeline with tests

System Architecture
High-Level Flow
text
Client                          Backend                      Database
  |                               |                             |
  |-- Master Password ----------->|                             |
  |   (derives encryption key)    |                             |
  |                               |                             |
  |-- Encrypted Vault Item ------>|-- Store encrypted blob --->|
  |   + Metadata                  |   + metadata                |
  |                               |                             |
  |<-- Encrypted blob ------------|<-- Fetch encrypted data ---|
  |   (decrypt client-side)       |                             |

Encryption Model
Master Password: User-provided, never sent to server
Master Key: Derived from master password using PBKDF2/Argon2
Encryption Key: AES-256-GCM key derived from master key
Vault Items: Encrypted client-side before transmission
Server Storage: Only encrypted blobs + unencrypted metadata (title, tags, created_at)
Security Boundaries
Server authenticates users but cannot decrypt vault contents
Encryption/decryption happens entirely on client
Server validates encrypted blob format but not contents
Recovery codes are hashed and stored securely
Audit logs track who accessed what, when

Database Schema
Users Table
sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255)
);
Sessions Table
sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);
Vault Items Table
sql
CREATE TABLE vault_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    encrypted_data BYTEA NOT NULL,
    item_type VARCHAR(50) NOT NULL, -- login, note, card, identity
    title VARCHAR(255),
    folder VARCHAR(255),
    tags TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    version INT DEFAULT 1
);
Shared Vault Items Table
sql
CREATE TABLE shared_vault_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_item_id UUID REFERENCES vault_items(id) ON DELETE CASCADE,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    encrypted_item_key BYTEA NOT NULL,
    permission VARCHAR(20) NOT NULL, -- read, write
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
Recovery Codes Table
sql
CREATE TABLE recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
Audit Log Table
sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

API Endpoints
Authentication
POST /api/v1/auth/signup - Create new account
POST /api/v1/auth/login - Login and receive token
POST /api/v1/auth/logout - Revoke current session
POST /api/v1/auth/refresh - Refresh auth token
GET /api/v1/auth/sessions - List active sessions
DELETE /api/v1/auth/sessions/{id} - Revoke specific session
MFA
POST /api/v1/mfa/enable - Enable TOTP MFA
POST /api/v1/mfa/verify - Verify TOTP code
POST /api/v1/mfa/disable - Disable MFA
Recovery
POST /api/v1/recovery/generate - Generate recovery codes
POST /api/v1/recovery/verify - Use recovery code to login
Vault Items
POST /api/v1/vault/items - Create vault item
GET /api/v1/vault/items - List vault items (with pagination)
GET /api/v1/vault/items/{id} - Get single vault item
PUT /api/v1/vault/items/{id} - Update vault item
DELETE /api/v1/vault/items/{id} - Soft delete vault item
POST /api/v1/vault/items/{id}/restore - Restore deleted item
Sharing
POST /api/v1/vault/items/{id}/share - Share item with user
DELETE /api/v1/vault/items/{id}/share/{user_id} - Revoke sharing
GET /api/v1/vault/shared - List items shared with me
Audit
GET /api/v1/audit/logs - Get audit logs for current user
Health
GET /health - Health check
GET /metrics - Prometheus metrics

Go Package Structure
text
vault-api/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth.go            # Auth handlers
│   │   │   ├── vault.go           # Vault handlers
│   │   │   ├── mfa.go             # MFA handlers
│   │   │   ├── recovery.go        # Recovery handlers
│   │   │   └── audit.go           # Audit handlers
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT/session validation
│   │   │   ├── ratelimit.go       # Rate limiting
│   │   │   ├── logging.go         # Request logging
│   │   │   └── recovery.go        # Panic recovery
│   │   └── router.go              # Route setup
│   ├── domain/
│   │   ├── user.go                # User entity
│   │   ├── vault_item.go          # Vault item entity
│   │   ├── session.go             # Session entity
│   │   └── audit_log.go           # Audit log entity
│   ├── service/
│   │   ├── auth_service.go        # Auth business logic
│   │   ├── vault_service.go       # Vault business logic
│   │   ├── mfa_service.go         # MFA logic
│   │   ├── recovery_service.go    # Recovery logic
│   │   └── audit_service.go       # Audit logging
│   ├── repository/
│   │   ├── user_repo.go           # User DB operations
│   │   ├── vault_repo.go          # Vault DB operations
│   │   ├── session_repo.go        # Session DB operations
│   │   └── audit_repo.go          # Audit DB operations
│   ├── crypto/
│   │   ├── hash.go                # Password hashing
│   │   ├── token.go               # Token generation
│   │   └── validation.go          # Input validation
│   └── config/
│       └── config.go              # Config management
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_sessions.sql
│   ├── 003_create_vault_items.sql
│   └── ...
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions CI
├── docs/
│   ├── architecture.md
│   ├── threat_model.md
│   └── openapi.yaml               # API spec
├── go.mod
├── go.sum
└── README.md

Tech Stack Details
Backend
Language: Go 1.22+
Router: chi or net/http
Database: PostgreSQL 16
Cache/Sessions: Redis 7
Migrations: golang-migrate or goose
DB Driver: pgx/v5
Security
Password Hashing: Argon2id or bcrypt
Encryption: AES-256-GCM (client-side)
JWT: golang-jwt/jwt
MFA: TOTP via pquerna/otp
Rate Limiting: golang.org/x/time/rate or Redis-based
DevOps
Containerization: Docker + Docker Compose
CI/CD: GitHub Actions
Logging: zerolog or logrus (JSON format)
Metrics: Prometheus + Grafana
Testing: testify for assertions, testcontainers-go for integration tests

Build Timeline (4-6 weeks)
Week 1: Foundation
Project setup (Go modules, Docker, PostgreSQL)
Database schema and migrations
User signup/login endpoints
Password hashing and JWT auth
Basic middleware (logging, CORS, auth)
Week 2: Vault Core
Vault item CRUD endpoints
Session management
Soft delete and restore
Pagination and filtering
Integration tests for vault operations
Week 3: Security Features
MFA (TOTP) implementation
Recovery codes
Rate limiting
Audit logging
Encrypted blob validation
Week 4: Sharing & Advanced
Vault item sharing between users
Key re-wrapping for password changes
Breach-aware password checks
OpenAPI documentation
More comprehensive tests
Week 5: Operations & Polish
Prometheus metrics
Health checks and readiness probes
CI/CD pipeline
Docker Compose for local dev
README, architecture docs, threat model
Week 6 (Optional): Demo Client
Simple CLI or web client
Demonstrates client-side encryption
Shows key derivation from master password
End-to-end flow demo

What Makes This Project Stand Out
1. Security-First Design
Zero-knowledge architecture
Threat model documentation
Proper cryptographic hygiene
Audit logging
2. Production-Ready
Dockerized deployment
CI/CD pipeline
Structured logging and metrics
Database migrations
3. Clean Architecture
Clear separation of concerns (handlers, services, repositories)
Dependency injection
Testable design
OpenAPI spec
4. Domain Relevance
Directly applicable to password management companies
Shows understanding of security boundaries
Demonstrates backend engineering maturity
5. Documentation
Architecture diagrams
Threat model
API documentation
Deployment guide
Design decision explanations

Threat Model (Key Points)
Threats Mitigated
Server compromise: Server cannot decrypt vault contents
MITM attacks: TLS required, encrypted payloads
Brute force: Rate limiting, strong password requirements
Session hijacking: Secure token storage, device tracking
Account takeover: MFA, recovery codes, audit logs
Threats Out of Scope
Client-side malware (assumes trusted client)
Physical device compromise
Social engineering of recovery codes
Quantum computing attacks on encryption
Security Assumptions
Users choose strong master passwords
Clients properly implement encryption
TLS properly configured
Server infrastructure is hardened

Success Metrics
For Interviews
Clear explanation of zero-knowledge architecture
Ability to walk through threat model
Discussion of trade-offs (security vs. UX vs. performance)
Demonstration of production-ready code
Technical Metrics
80%+ test coverage
Sub-100ms p95 latency for vault operations
Clean CI pipeline (linting, tests pass)
Working Docker deployment
Complete OpenAPI documentation

Next Steps
Set up repository with Go modules and Docker
Implement auth flow (signup, login, sessions)
Build vault CRUD with encrypted storage
Add security features (MFA, recovery, audit logs)
Polish operations (metrics, CI/CD, docs)
Create demo (CLI or web client showing encryption)

Additional Resources
Zero-knowledge architecture: https://bitwarden.com/help/bitwarden-security-white-paper/
Go project layout: https://github.com/golang-standards/project-layout
Secure password storage: OWASP Password Storage Cheat Sheet
TOTP specification: RFC 6238

