-- name: wallet_get :one
-- Get wallet for a project (scoped by user)
SELECT id, created_at, updated_at, project_id, user_id, address, chain_id
FROM project_wallets
WHERE user_id = $1 AND project_id = $2;
