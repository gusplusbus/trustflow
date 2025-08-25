-- +goose Up
-- Creates a single-wallet-per-project table with FK to projects.

CREATE TABLE IF NOT EXISTS project_wallets (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     text NOT NULL,

    address     text NOT NULL,
    chain_id    integer NOT NULL,

    UNIQUE (project_id)
);

CREATE INDEX IF NOT EXISTS idx_project_wallets_user ON project_wallets(user_id);

-- Trigger function to maintain updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION trg_project_wallets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Ensure single trigger is present
DROP TRIGGER IF EXISTS set_timestamp_project_wallets ON project_wallets;

CREATE TRIGGER set_timestamp_project_wallets
BEFORE UPDATE ON project_wallets
FOR EACH ROW EXECUTE FUNCTION trg_project_wallets_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS set_timestamp_project_wallets ON project_wallets;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS trg_project_wallets_updated_at();
-- +goose StatementEnd

DROP INDEX IF EXISTS idx_project_wallets_user;
DROP TABLE IF EXISTS project_wallets;
