SELECT
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  gh_issue_id, gh_number,
  title, state, html_url,
  labels, user_login AS gh_user_login,
  gh_created_at, gh_updated_at
FROM project_issues
WHERE user_id = $1
  AND project_id    = $2
ORDER BY gh_number ASC;
