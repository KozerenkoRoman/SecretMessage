-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, user_role)
VALUES ($1, $2, $3, 'user')
RETURNING id, username, email, user_role, is_banned, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, password_hash, user_role, is_banned, ban_reason, banned_at, created_at, updated_at
FROM users 
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, user_role, is_banned, ban_reason, banned_at, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, user_role, is_banned, ban_reason, banned_at, created_at, updated_at
FROM users
WHERE username = $1;

-- name: BanUser :exec
UPDATE users 
SET is_banned = TRUE, ban_reason = $2, banned_at = NOW(), updated_at = NOW() 
WHERE id = $1;

-- name: UnbanUser :exec
UPDATE users 
SET is_banned = FALSE, ban_reason = NULL, banned_at = NULL, updated_at = NOW() 
WHERE id = $1;

-- name: UpdateUserStats :exec
INSERT INTO user_stats (user_id, games_played, games_won, spy_bonuses_received, total_score, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (user_id) 
DO UPDATE SET 
    games_played = user_stats.games_played + EXCLUDED.games_played,
    games_won = user_stats.games_won + EXCLUDED.games_won,
    spy_bonuses_received = user_stats.spy_bonuses_received + EXCLUDED.spy_bonuses_received,
    total_score = user_stats.total_score + EXCLUDED.total_score,
    updated_at = NOW();

-- name: GetUserStats :one
SELECT user_id, games_played, games_won, spy_bonuses_received, total_score, updated_at 
FROM user_stats 
WHERE user_id = $1;

-- name: LogGameHistory :exec
INSERT INTO game_history (room_id, winner_id, final_state, played_at)
VALUES ($1, $2, $3, NOW());

-- name: GetUserGameHistory :many
SELECT id, room_id, winner_id, final_state, played_at 
FROM game_history 
WHERE winner_id = $1 
ORDER BY played_at DESC 
LIMIT $2 OFFSET $3;

-- name: SeedAdminUser :exec
INSERT INTO users (username, email, password_hash, user_role)
VALUES ($1, $2, $3, 'admin')
ON CONFLICT (username) DO NOTHING;

-- name: GetLeaderboard :many
SELECT u.username, s.games_played, s.games_won, s.total_score, s.updated_at
FROM user_stats as s
JOIN users as u ON s.user_id = u.id
ORDER BY s.total_score DESC, s.games_won DESC
LIMIT $1;