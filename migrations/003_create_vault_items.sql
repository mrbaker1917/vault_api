-- +goose Up
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

CREATE TABLE shared_vault_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_item_id UUID REFERENCES vault_items(id) ON DELETE CASCADE,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    encrypted_item_key BYTEA NOT NULL,
    permission VARCHAR(20) NOT NULL, -- read, write
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

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

-- +goose Down
DROP TABLE IF EXISTS shared_vault_items;
DROP TABLE IF EXISTS recovery_codes;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS vault_items;
