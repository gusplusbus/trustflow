-- Fetch leaves in order (for proofs or to rebuild root)
-- Params: $1 entity_kind, $2 entity_key, $3 bucket_key
SELECT leaf_index, leaf_hash
FROM timeline_bucket_leaves
WHERE entity_kind = $1 AND entity_key = $2 AND bucket_key = $3
ORDER BY leaf_index ASC;
