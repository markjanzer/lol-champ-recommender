-- name: LogPlayerSearch :exec
INSERT INTO player_search_log (player_id, search_time) VALUES ($1, $2) RETURNING id;

-- name: FindPlayerSearchLog :one
SELECT * FROM player_search_log WHERE player_id = $1 ORDER BY search_time DESC LIMIT 1;