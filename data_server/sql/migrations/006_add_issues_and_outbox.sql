-- +goose Up
CREATE TABLE IF NOT EXISTS project_issues (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),

  user_id    uuid NOT NULL,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

  organization text NOT NULL,
  repository   text NOT NULL,

  gh_issue_id  bigint NOT NULL,
  gh_number    integer NOT NULL,

  title        text NOT NULL,
  state        text NOT NULL,
  html_url     text NOT NULL,
  user_login   text DEFAULT '',
  labels       text[] DEFAULT '{}',

  gh_created_at timestamptz,
  gh_updated_at timestamptz
);

-- all queries are scoped by user/project; this helps listing
CREATE INDEX IF NOT EXISTS idx_project_issues_project ON project_issues(project_id);
CREATE INDEX IF NOT EXISTS idx_project_issues_user_project ON project_issues(user_id, project_id);

-- idempotency key: never import the same GH issue twice into the same project
CREATE UNIQUE INDEX IF NOT EXISTS ux_project_issues_project_gh
  ON project_issues(project_id, gh_issue_id);

-- Outbox for the ledger worker (generic)
CREATE TABLE IF NOT EXISTS issue_outbox (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  event_type text NOT NULL,         -- e.g. "project_issue.imported.v1"
  payload jsonb NOT NULL,           -- full issue row payload
  published_at timestamptz NULL,    -- set by worker
  attempts int NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_issue_outbox_published ON issue_outbox(published_at NULLS FIRST);

-- +goose Down
DROP TABLE IF EXISTS issue_outbox;
DROP INDEX IF EXISTS ux_project_issues_project_gh;
DROP INDEX IF EXISTS idx_project_issues_user_project;
DROP INDEX IF EXISTS idx_project_issues_project;
DROP TABLE IF EXISTS project_issues;
