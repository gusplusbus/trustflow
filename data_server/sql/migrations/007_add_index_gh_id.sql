-- +goose Up

CREATE INDEX IF NOT EXISTS idx_project_issues_gh_issue_id
  ON project_issues (gh_issue_id);

-- +goose Down
DROP INDEX IF EXISTS idx_project_issues_gh_issue_id;
