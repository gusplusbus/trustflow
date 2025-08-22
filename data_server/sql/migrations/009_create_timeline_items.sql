-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS timeline_items (
  id                BIGSERIAL PRIMARY KEY,
  entity_kind       TEXT NOT NULL,              -- 'issue' | 'pr' | 'sub_issue' (future)
  entity_key        TEXT NOT NULL,              -- 'owner/repo#123'
  provider          TEXT NOT NULL,              -- 'github'
  provider_event_id TEXT NOT NULL,              -- GraphQL node id (unique)
  type              TEXT NOT NULL,              -- IssueComment, LabeledEvent, ...
  actor             TEXT,
  created_at        TIMESTAMPTZ NOT NULL,
  payload_json      JSONB NOT NULL,
  item_hash         BYTEA NOT NULL,             -- hash(DAG-CBOR(canonical item))
  bucket_key        TEXT NOT NULL,              -- e.g. '2025-08-22'
  seq_in_entity     BIGINT NOT NULL,            -- per-entity monotonic sequence
  UNIQUE (provider_event_id)
);

-- range/time + bucket scans
CREATE INDEX IF NOT EXISTS timeline_items_entity_created_idx
  ON timeline_items (entity_kind, entity_key, created_at);

CREATE INDEX IF NOT EXISTS timeline_items_entity_bucket_idx
  ON timeline_items (entity_kind, entity_key, bucket_key);

-- quick lookup by provider_event_id (unique already, but index helps joins)
CREATE INDEX IF NOT EXISTS timeline_items_provider_event_idx
  ON timeline_items (provider_event_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS timeline_items_provider_event_idx;
DROP INDEX IF EXISTS timeline_items_entity_bucket_idx;
DROP INDEX IF EXISTS timeline_items_entity_created_idx;
DROP TABLE IF EXISTS timeline_items;
-- +goose StatementEnd
