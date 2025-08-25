-- name: wallet_upsert :one
-- Insert or update the wallet for a project
INSERT INTO project_wallets (project_id, user_id, address, chain_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (project_id) DO UPDATE
SET address = EXCLUDED.address,
chain_id = EXCLUDED.chain_id,
updated_at = now()
RETURNING id, created_at, updated_at, project_id, user_id, address, chain_id;
