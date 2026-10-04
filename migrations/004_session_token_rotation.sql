-- +goose Up
ALTER TABLE sessions
    ADD COLUMN previous_token_hash VARCHAR(255),
    ADD COLUMN token_rotated_at TIMESTAMP;

-- +goose Down
ALTER TABLE sessions
    DROP COLUMN IF EXISTS token_rotated_at,
    DROP COLUMN IF EXISTS previous_token_hash;
