-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  user_id UUID NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  duration_estimate INTEGER NOT NULL,
  team_size INTEGER NOT NULL,
  application_close_time TEXT
);

-- +goose Down
DROP TABLE IF EXISTS projects;
