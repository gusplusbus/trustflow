WITH filtered AS (
  SELECT
    id,
    created_at,
    updated_at,
    title,
    description,
    duration_estimate,
    team_size,
    application_close_time
  FROM projects
  WHERE user_id = $1
    AND (
      $2 = '' OR
      title ILIKE '%%' || $2 || '%%' OR
      description ILIKE '%%' || $2 || '%%'
    )
)
SELECT *
FROM filtered
ORDER BY %s %s
OFFSET $3
LIMIT $4;
