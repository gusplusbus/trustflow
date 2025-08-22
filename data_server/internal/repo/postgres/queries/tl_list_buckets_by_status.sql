-- List buckets by status (e.g., 'needs_anchoring'), newest first
-- Params: $1 status TEXT, $2 limit INT, $3 offset INT
SELECT entity_kind, entity_key, bucket_key,
       root_hash, leaf_count, status, cid, closed_at, anchored_tx, anchored_at
FROM timeline_buckets
WHERE status = $1
ORDER BY bucket_key DESC
LIMIT $2 OFFSET $3;
