-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS timeline_buckets (
  entity_kind  TEXT NOT NULL,
  entity_key   TEXT NOT NULL,
  bucket_key   TEXT NOT NULL,
  root_hash    BYTEA NOT NULL,              -- current/final Merkle root
  leaf_count   INT    NOT NULL DEFAULT 0,
  status       TEXT   NOT NULL DEFAULT 'open',   -- open|closed|needs_anchoring|anchored
  cid          TEXT,
  closed_at    TIMESTAMPTZ,
  anchored_tx  TEXT,
  anchored_at  TIMESTAMPTZ,
  PRIMARY KEY (entity_kind, entity_key, bucket_key)
);

-- helpful listing/paging
CREATE INDEX IF NOT EXISTS timeline_buckets_scope_idx
  ON timeline_buckets (entity_kind, entity_key, bucket_key DESC);

CREATE INDEX IF NOT EXISTS timeline_buckets_status_idx
  ON timeline_buckets (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS timeline_buckets_status_idx;
DROP INDEX IF EXISTS timeline_buckets_scope_idx;
DROP TABLE IF EXISTS timeline_buckets;
-- +goose StatementEnd
