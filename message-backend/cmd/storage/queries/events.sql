-- name: InsertGameEvent :one
INSERT INTO game_events (
    room_id, 
    turn_id, 
    event_id, 
    event_type, 
    payload, 
    state_before
) 
SELECT 
    $1, 
    $2, 
    (SELECT COUNT(*) FROM game_events WHERE room_id = $1) + 1, 
    $3, 
    $4, 
    $5
RETURNING event_id;

-- name: GetGameEventsByRoom :many
SELECT sqlc.embed(game_events)
FROM game_events
WHERE room_id = $1
ORDER BY event_id ASC;

-- name: DeleteGameEventsByRoom :exec
DELETE FROM game_events
WHERE room_id = $1;