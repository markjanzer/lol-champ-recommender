-- name: LogPlayerSearch :exec
INSERT INTO player_search_log (player_id) VALUES ($1) RETURNING id;

-- name: PlayerHasBeenSearched :one
SELECT EXISTS(SELECT 1 FROM player_search_log WHERE player_id = $1);

-- name: LastSearched :one
SELECT search_time FROM player_search_log WHERE player_id = $1 ORDER BY search_time DESC LIMIT 1;