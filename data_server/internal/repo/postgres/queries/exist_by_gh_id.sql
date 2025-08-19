SELECT EXISTS(SELECT 1 FROM project_issues WHERE gh_issue_id = $1);
