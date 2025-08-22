-- Insert a leaf row (precomputed leaf_index), ignore duplicates
-- Params:
--   $1 entity_kind TEXT
--   $2 entity_key  TEXT
--   $3 bucket_key  TEXT
--   $4 leaf_index  INT
--   $5 leaf_hash   BYTEA
INSERT INTO timeline_bucket_leaves (entity_kind, entity_key, bucket_key, leaf_index, leaf_hash)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

