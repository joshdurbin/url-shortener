-- name: CreateURL :one
INSERT INTO urls (short_code, original_url, created_at, usage_count)
VALUES (?, ?, ?, 0)
RETURNING *;

-- name: GetURL :one
SELECT * FROM urls
WHERE short_code = ?;

-- name: GetAllURLs :many
SELECT * FROM urls
ORDER BY created_at DESC;

-- name: UpdateUsage :exec
UPDATE urls 
SET usage_count = ?, last_used_at = ?
WHERE short_code = ?;

-- name: DeleteURL :exec
DELETE FROM urls 
WHERE short_code = ?;

-- name: URLExists :one
SELECT COUNT(*) FROM urls
WHERE short_code = ?;