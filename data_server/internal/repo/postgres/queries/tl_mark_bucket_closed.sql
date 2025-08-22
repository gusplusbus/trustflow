-- Transition open -> needs_anchoring and stamp closed_at once
-- Params: $1 entity_kind, $2 entity_key, $3 bucket_key
UPDATE timeline_buckets
SET status = 'needs_anchoring',
    closed_at = COALESCE(closed_at, now())
WHERE entity_kind = $1 AND entity_key = $2 AND bucket_key = $3
  AND status = 'open'
RETURNING entity_kind, entity_key, bucket_key,
          root_hash, leaf_count, status, cid, closed_at, anchored_tx, anchored_at;
