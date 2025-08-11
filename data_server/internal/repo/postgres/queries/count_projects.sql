SELECT COUNT(*)
FROM projects
WHERE user_id = $1
  AND (
    $2 = '' OR
    title ILIKE '%' || $2 || '%' OR
    description ILIKE '%' || $2 || '%'
  );
