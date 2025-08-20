-- name: UpsertIssuesTimelineCheckpoint :exec
INSERT INTO issues_timeline_checkpoint (project_issue_id, cursor, last_event_at, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (project_issue_id)
DO UPDATE SET cursor = EXCLUDED.cursor,
              last_event_at = EXCLUDED.last_event_at,
              updated_at = now();
