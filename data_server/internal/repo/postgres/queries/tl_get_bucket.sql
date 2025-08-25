-- Get a single bucket row
-- Params: $1 entity_kind, $2 entity_key, $3 bucket_key
SELECT entity_kind, entity_key, bucket_key,
       root_hash, leaf_count, status, cid, closed_at, anchored_tx, anchored_at
FROM timeline_buckets
WHERE entity_kind = $1 AND entity_key = $2 AND bucket_key = $3;
