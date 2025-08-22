-- Insert one timeline item (idempotent on provider_event_id)
-- Params:
--   $1 entity_kind TEXT
--   $2 entity_key  TEXT
--   $3 provider    TEXT
--   $4 provider_event_id TEXT
--   $5 type        TEXT
--   $6 actor       TEXT NULL
--   $7 created_at  TIMESTAMPTZ
--   $8 payload_json JSONB
--   $9 item_hash   BYTEA
--   $10 bucket_key TEXT
INSERT INTO timeline_items (
  entity_kind, entity_key, provider, provider_event_id,
  type, actor, created_at, payload_json, item_hash, bucket_key, seq_in_entity
)
VALUES (
  $1, $2, $3, $4,
  $5, $6, $7, $8, $9, $10,
  COALESCE((
    SELECT MAX(seq_in_entity)+1
    FROM timeline_items
    WHERE entity_kind = $1 AND entity_key = $2
  ), 1)
)
ON CONFLICT (provider_event_id) DO NOTHING;
