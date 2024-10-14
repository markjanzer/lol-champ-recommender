-- name: CreateChampionStats :exec
INSERT INTO champion_stats (data) VALUES ($1);

-- name: GetLastChampionStats :one
SELECT * FROM champion_stats ORDER BY created_at DESC LIMIT 1;

-- name: AllChampions :many
SELECT * FROM champions;
