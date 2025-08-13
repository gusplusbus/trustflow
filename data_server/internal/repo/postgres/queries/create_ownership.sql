INSERT INTO ownerships (
  id, created_at, updated_at,
  user_id, project_id,
  organization, repository,
  provider, web_url
)
VALUES (
  gen_random_uuid(),
  now() AT TIME ZONE 'utc',
  now() AT TIME ZONE 'utc',
  $1, $2,
  $3, $4,
  NULLIF($5, ''), NULLIF($6, '')
)
RETURNING
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  COALESCE(provider, ''), COALESCE(web_url, '');
