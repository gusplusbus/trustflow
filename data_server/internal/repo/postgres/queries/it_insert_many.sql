-- name: InsertManyIssuesTimelineRaw :copyfrom
INSERT INTO issues_timeline_raw (
  project_issue_id, provider, provider_event_id, type, actor, created_at, payload_json
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (project_issue_id, provider_event_id) DO NOTHING;
