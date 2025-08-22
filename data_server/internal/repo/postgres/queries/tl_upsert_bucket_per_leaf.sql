-- Upsert/advance a bucket by ONE appended leaf (root precomputed in app)
-- Params:
--   $1 entity_kind TEXT
--   $2 entity_key  TEXT
--   $3 bucket_key  TEXT
--   $4 new_root    BYTEA
INSERT INTO timeline_buckets (entity_kind, entity_key, bucket_key, root_hash, leaf_count, status)
VALUES ($1, $2, $3, $4, 1, 'open')
ON CONFLICT (entity_kind, entity_key, bucket_key)
DO UPDATE SET
  root_hash = EXCLUDED.root_hash,
  leaf_count = timeline_buckets.leaf_count + 1;
