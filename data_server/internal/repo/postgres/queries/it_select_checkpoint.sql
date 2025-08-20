-- name: GetIssuesTimelineCheckpoint :one
SELECT cursor, last_event_at, updated_at
FROM issues_timeline_checkpoint
WHERE project_issue_id = $1;
