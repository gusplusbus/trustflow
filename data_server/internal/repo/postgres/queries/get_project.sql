SELECT id, created_at, updated_at, title, description,
       duration_estimate, team_size, application_close_time
FROM projects
WHERE id = $1 AND user_id = $2;
