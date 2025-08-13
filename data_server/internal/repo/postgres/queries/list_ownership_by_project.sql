SELECT
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  COALESCE(provider, ''), COALESCE(web_url, '')
FROM ownerships
WHERE user_id = $1
  AND project_id = $2
ORDER BY created_at DESC;
