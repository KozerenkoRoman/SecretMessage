-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, user_role, avatar_seed)
VALUES ($1, $2, $3, 'user', $4)
RETURNING id, username, email, user_role, is_banned, avatar_seed, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, password_hash, user_role, is_banned, ban_reason, banned_at, avatar_seed, created_at, updated_at
FROM users 
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, user_role, is_banned, ban_reason, banned_at, avatar_seed, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, user_role, is_banned, ban_reason, banned_at, avatar_seed, created_at, updated_at
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

-- name: SeedAdminUser :exec
INSERT INTO users (username, email, password_hash, user_role, avatar_seed)
VALUES ($1, $2, $3, 'admin', $4)
ON CONFLICT (username) DO NOTHING;

-- name: SeedBotUser :exec
INSERT INTO users (id, username, email, password_hash, user_role, avatar_seed)
VALUES ($1, $2, $3, $4, 'bot', $5)
ON CONFLICT (username) DO NOTHING;

-- name: GetLeaderboard :many
SELECT u.username, u.user_role, u.avatar_seed, sqlc.embed(s)
FROM user_stats as s
JOIN users as u ON s.user_id = u.id
ORDER BY s.total_score DESC, s.games_won DESC
LIMIT $1;

-- name: UpdateUser :exec
UPDATE users 
SET avatar_seed = $2, 
username = $3,
email = $4,
updated_at = NOW() 
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2, 
updated_at = NOW() 
WHERE id = $1;

-- name: UpdateUserStats :exec
INSERT INTO user_stats (user_id, games_played, games_won, rounds_played, rounds_won, spy_bonuses, total_score, updated_at) 
VALUES (
    $1, $2, $3, $4, $5, $6, $7, now()
)
ON CONFLICT (user_id) DO UPDATE SET
    games_played = user_stats.games_played + EXCLUDED.games_played,
    games_won = user_stats.games_won + EXCLUDED.games_won,
    rounds_played = user_stats.rounds_played + EXCLUDED.rounds_played,
    rounds_won = user_stats.rounds_won + EXCLUDED.rounds_won,
    spy_bonuses = user_stats.spy_bonuses + EXCLUDED.spy_bonuses,
    total_score = user_stats.total_score + EXCLUDED.total_score,
    updated_at = now();

-- name: GetUserStats :one
SELECT sqlc.embed(user_stats)
FROM user_stats 
WHERE user_id = $1;
