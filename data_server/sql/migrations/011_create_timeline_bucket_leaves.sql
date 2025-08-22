-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS timeline_bucket_leaves (
  entity_kind  TEXT NOT NULL,
  entity_key   TEXT NOT NULL,
  bucket_key   TEXT NOT NULL,
  leaf_index   INT  NOT NULL,
  leaf_hash    BYTEA NOT NULL,
  PRIMARY KEY (entity_kind, entity_key, bucket_key, leaf_index)
);

-- fetch by bucket in order
CREATE INDEX IF NOT EXISTS timeline_bucket_leaves_bucket_idx
  ON timeline_bucket_leaves (entity_kind, entity_key, bucket_key, leaf_index);

-- occasionally useful to look up by hash (low collision risk)
CREATE INDEX IF NOT EXISTS timeline_bucket_leaves_hash_idx
  ON timeline_bucket_leaves (entity_kind, entity_key, bucket_key, leaf_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS timeline_bucket_leaves_hash_idx;
DROP INDEX IF EXISTS timeline_bucket_leaves_bucket_idx;
DROP TABLE IF EXISTS timeline_bucket_leaves;
-- +goose StatementEnd
