INSERT INTO projects (
  id, created_at, updated_at, user_id, title, description,
  duration_estimate, team_size, application_close_time
)
VALUES (
  gen_random_uuid(), now() AT TIME ZONE 'utc', now() AT TIME ZONE 'utc',
  $1, $2, $3, $4, $5, $6
)
RETURNING id, created_at, updated_at, title, description,
          duration_estimate, team_size, application_close_time;
