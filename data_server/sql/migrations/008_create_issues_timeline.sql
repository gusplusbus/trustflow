-- +goose Up
-- +goose StatementBegin
/*
  Issues timeline (append-only) + per-issue checkpoint.

  - issues_timeline_raw:
      one row per timeline event fetched from GitHub GraphQL
      dedup by (project_issue_id, provider_event_id)

  - issues_timeline_checkpoint:
      resume token (GraphQL cursor) per issue
*/

CREATE TABLE IF NOT EXISTS issues_timeline_raw (
  id                 BIGSERIAL PRIMARY KEY,
  project_issue_id   UUID NOT NULL REFERENCES project_issues(id) ON DELETE CASCADE,
  provider           TEXT NOT NULL,                         -- e.g. 'github'
  provider_event_id  TEXT NOT NULL,                         -- GraphQL node id (unique per issue)
  type               TEXT NOT NULL,                         -- e.g. IssueComment, AssignedEvent, ...
  actor              TEXT,                                  -- login (nullable)
  created_at         TIMESTAMPTZ NOT NULL,                  -- from GH
  payload_json       JSONB NOT NULL,                        -- raw-ish event payload
  inserted_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (project_issue_id, provider_event_id)
);

-- helpful for range scans / ordering within an issue
CREATE INDEX IF NOT EXISTS issues_timeline_raw_issue_created_idx
  ON issues_timeline_raw(project_issue_id, created_at);

-- checkpoint per issue: last GraphQL cursor and last seen event time
CREATE TABLE IF NOT EXISTS issues_timeline_checkpoint (
  project_issue_id UUID PRIMARY KEY REFERENCES project_issues(id) ON DELETE CASCADE,
  cursor           TEXT NOT NULL DEFAULT '',
  last_event_at    TIMESTAMPTZ,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS issues_timeline_checkpoint;
DROP TABLE IF EXISTS issues_timeline_raw;
-- +goose StatementEnd
