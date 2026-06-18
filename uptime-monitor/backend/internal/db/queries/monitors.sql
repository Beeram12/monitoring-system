-- name: CreateMonitor :one
INSERT INTO monitors (url, name, interval_sec)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMonitor :one
SELECT * FROM monitors WHERE id = $1;

-- name: ListMonitors :many
SELECT * FROM monitors ORDER BY id ASC;

-- name: DeleteMonitor :exec
DELETE FROM monitors WHERE id = $1;
