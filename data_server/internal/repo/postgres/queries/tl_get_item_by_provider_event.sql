-- Find the bucket+hash for a specific provider_event_id (for InclusionProof)
-- Params: $1 provider_event_id
SELECT entity_kind, entity_key, bucket_key, seq_in_entity, item_hash
FROM timeline_items
WHERE provider_event_id = $1;
