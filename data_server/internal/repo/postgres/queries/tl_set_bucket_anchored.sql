-- Set CID + chain tx and mark as anchored
-- Params: $1 entity_kind, $2 entity_key, $3 bucket_key, $4 cid, $5 anchored_tx
UPDATE timeline_buckets
SET cid = $4,
    anchored_tx = $5,
    anchored_at = now(),
    status = 'anchored'
WHERE entity_kind = $1 AND entity_key = $2 AND bucket_key = $3
RETURNING entity_kind, entity_key, bucket_key,
          root_hash, leaf_count, status, cid, closed_at, anchored_tx, anchored_at;
