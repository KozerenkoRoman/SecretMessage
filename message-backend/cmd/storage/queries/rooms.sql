-- name: SaveRoom :exec
INSERT INTO rooms (id, state, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (id)
DO UPDATE SET state = EXCLUDED.state, updated_at = NOW();

-- name: GetRoom :one
SELECT id, state, updated_at FROM rooms WHERE id = $1;

-- name: DeleteRoom :exec
DELETE FROM rooms WHERE id = $1;

-- name: CheckHealth :one
SELECT 1 as ok;