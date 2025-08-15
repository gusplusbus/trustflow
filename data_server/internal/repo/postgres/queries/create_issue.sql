-- insert but DO NOT overwrite; when duplicate, no row is returned
INSERT INTO issues (
  project_id, user_id,
  organization, repository,
  gh_issue_id, gh_number,
  title, state, html_url,
  labels, gh_user_login,
  gh_created_at, gh_updated_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (project_id, organization, repository, gh_number)
DO NOTHING
RETURNING
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  gh_issue_id, gh_number,
  title, state, html_url,
  labels, gh_user_login,
  gh_created_at, gh_updated_at;
