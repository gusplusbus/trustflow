-- q_project_issues_insert_many.sql
-- Params (arrays; all must be same length):
--  $1  :: uuid[]        -- user_id
--  $2  :: uuid[]        -- project_id
--  $3  :: text[]        -- organization
--  $4  :: text[]        -- repository
--  $5  :: bigint[]      -- gh_issue_id
--  $6  :: int[]         -- gh_number
--  $7  :: text[]        -- title
--  $8  :: text[]        -- state
--  $9  :: text[]        -- html_url
--  $10 :: text[][]      -- labels (each row is text[])
--  $11 :: text[]        -- user_login
--  $12 :: timestamptz[] -- gh_created_at (nullable -> use NULLs if unknown)
--  $13 :: timestamptz[] -- gh_updated_at (nullable)
--  $14 :: timestamptz[] -- created_at
--  $15 :: timestamptz[] -- updated_at

WITH input AS (
  SELECT
    u, p, o, r, i, n,
    t, s, h,
    l, ul, gca, gua,
    ca, ua
  FROM unnest(
    $1::uuid[],        -- u  user_id
    $2::uuid[],        -- p  project_id
    $3::text[],        -- o  organization
    $4::text[],        -- r  repository
    $5::bigint[],      -- i  gh_issue_id
    $6::int[],         -- n  gh_number
    $7::text[],        -- t  title
    $8::text[],        -- s  state
    $9::text[],        -- h  html_url
    $10::text[][],     -- l  labels (row-wise text[])
    $11::text[],       -- ul user_login
    $12::timestamptz[],-- gca gh_created_at
    $13::timestamptz[],-- gua gh_updated_at
    $14::timestamptz[],-- ca  created_at
    $15::timestamptz[] -- ua  updated_at
  ) AS x(u,p,o,r,i,n,t,s,h,l,ul,gca,gua,ca,ua)
),
ins AS (
  INSERT INTO project_issues (
    user_id, project_id,
    organization, repository,
    gh_issue_id, gh_number,
    title, state, html_url,
    user_login, labels,
    gh_created_at, gh_updated_at,
    created_at, updated_at
  )
  SELECT
    u, p,
    o, r,
    i, n,
    t, s, h,
    ul, COALESCE(l, '{}')::text[],
    gca, gua,
    ca, ua
  FROM input
  ON CONFLICT (project_id, gh_issue_id)
  DO NOTHING
  RETURNING
    id, created_at, updated_at,
    project_id, user_id,
    organization, repository,
    gh_issue_id, gh_number,
    title, state, html_url,
    labels, user_login,
    gh_created_at, gh_updated_at
)
SELECT
  id, created_at, updated_at,
  project_id, user_id,
  organization, repository,
  gh_issue_id, gh_number,
  title, state, html_url,
  labels, user_login,
  gh_created_at, gh_updated_at
FROM ins
ORDER BY gh_number ASC;
