UPDATE ownerships
SET
  organization = $1,
  repository   = $2,
  provider     = NULLIF($3, ''),
  web_url      = NULLIF($4, ''),
  updated_at   = now() AT TIME ZONE 'utc'
WHERE id = $5 AND user_id = $6
RETURNING
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  COALESCE(provider, ''), COALESCE(web_url, '');
