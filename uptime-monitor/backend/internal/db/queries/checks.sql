-- name: CreateCheck :one
INSERT INTO checks (monitor_id, status_code, response_ms, is_up, error)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListChecksByMonitor :many
SELECT * FROM checks
WHERE monitor_id = $1
ORDER BY checked_at DESC
LIMIT $2;

-- name: GetLatestCheckByMonitor :one
SELECT * FROM checks
WHERE monitor_id = $1
ORDER BY checked_at DESC
LIMIT 1;

-- name: GetLatestChecksForAllMonitors :many
SELECT DISTINCT ON (monitor_id) *
FROM checks
ORDER BY monitor_id, checked_at DESC;
