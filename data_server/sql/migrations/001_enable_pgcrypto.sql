-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose Down
-- (No-op: leaving pgcrypto installed is fine)

