-- +goose Up
CREATE TABLE IF NOT EXISTS ownerships (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id     TEXT NOT NULL,

  organization TEXT NOT NULL,
  repository   TEXT NOT NULL,

  provider     TEXT,
  web_url      TEXT
);

-- one org/repo per project
CREATE UNIQUE INDEX IF NOT EXISTS ux_ownerships_project_org_repo
  ON ownerships (project_id, organization, repository);

-- common lookup
CREATE INDEX IF NOT EXISTS ix_ownerships_user_project
  ON ownerships (user_id, project_id);

-- keep updated_at fresh (wrap in StatementBegin/End for goose)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_ownerships_updated_at ON ownerships;

CREATE TRIGGER trg_ownerships_updated_at
BEFORE UPDATE ON ownerships
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_ownerships_updated_at ON ownerships;
DROP TABLE IF EXISTS ownerships;
