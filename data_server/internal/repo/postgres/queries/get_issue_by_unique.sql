SELECT
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  gh_issue_id, gh_number,
  title, state, html_url,
  labels, gh_user_login,
  gh_created_at, gh_updated_at
FROM issues
WHERE project_id = $1
  AND organization = $2
  AND repository   = $3
  AND gh_number    = $4;
