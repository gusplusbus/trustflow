UPDATE projects
SET title=$1,
    description=$2,
    duration_estimate=$3,
    team_size=$4,
    application_close_time=$5,
    updated_at=now() AT TIME ZONE 'utc'
WHERE id=$6 AND user_id=$7
RETURNING id, created_at, updated_at, title, description,
          duration_estimate, team_size, application_close_time;
