-- List buckets for a scope with basic pagination
-- Params:
--   $1 entity_kind, $2 entity_key, $3 limit, $4 offset
SELECT entity_kind, entity_key, bucket_key,
       root_hash, leaf_count, status, cid, closed_at, anchored_tx, anchored_at
FROM timeline_buckets
WHERE entity_kind = $1 AND entity_key = $2
ORDER BY bucket_key DESC
LIMIT $3 OFFSET $4;
