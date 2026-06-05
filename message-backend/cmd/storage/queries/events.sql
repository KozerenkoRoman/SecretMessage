-- name: LogTurn :one
INSERT INTO game_turns (
    room_id, sequence_id, action_type, player_id, hand_index, target_id, guess_card, played_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
RETURNING id;

-- name: LogDomainEvent :exec
INSERT INTO game_events (room_id, turn_id, event_id, event_type, payload, created_at)
VALUES ($1, $2, $3, $4, $5, NOW());

-- name: GetRoomHistory :many
SELECT 
    t.sequence_id, t.action_type, t.player_id, t.hand_index, t.target_id, t.guess_card, t.played_at,
    e.event_id, e.event_type, e.payload
FROM game_turns t
LEFT JOIN game_events e ON t.id = e.turn_id
WHERE t.room_id = $1
ORDER BY t.sequence_id ASC, e.event_id ASC;