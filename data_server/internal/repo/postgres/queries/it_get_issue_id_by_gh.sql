-- name: GetProjectIssueIDByGhID :one
SELECT id
FROM project_issues
WHERE gh_issue_id = $1
LIMIT 1;

