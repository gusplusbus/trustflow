-- +goose Up
-- Enable pg_trgm for trigram indexing (idempotent)
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

-- Composite index for: WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_projects_user_created
    ON projects (user_id, created_at DESC);

-- Trigram indexes to speed up ILIKE searches on title/description
CREATE INDEX IF NOT EXISTS idx_projects_title_trgm
    ON projects USING GIN (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_projects_desc_trgm
    ON projects USING GIN (description gin_trgm_ops);


-- +goose Down
-- Drop trigram indexes first (they depend on pg_trgm)
DROP INDEX IF EXISTS idx_projects_desc_trgm;
DROP INDEX IF EXISTS idx_projects_title_trgm;

-- Drop the composite index
DROP INDEX IF EXISTS idx_projects_user_created;

-- Finally, remove the extension (only if nothing else needs it)
DROP EXTENSION IF EXISTS pg_trgm;
