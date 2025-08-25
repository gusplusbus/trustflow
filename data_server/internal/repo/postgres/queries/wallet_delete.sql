-- name: wallet_delete :execrows
-- Delete the wallet for a project
DELETE FROM project_wallets WHERE user_id = $1 AND project_id = $2;
