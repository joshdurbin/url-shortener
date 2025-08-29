-- name: GetCounter :one
SELECT value FROM counters WHERE key = ?;

-- name: SetCounter :exec
INSERT OR REPLACE INTO counters (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP);

-- name: IncrementCounter :one
INSERT INTO counters (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET 
    value = counters.value + ?,
    updated_at = CURRENT_TIMESTAMP
RETURNING value;