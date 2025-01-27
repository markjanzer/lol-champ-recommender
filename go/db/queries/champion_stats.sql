-- name: CreateChampionStats :exec
INSERT INTO champion_stats (
  data,
  last_match_id
) VALUES ($1, $2);

-- name: LastChampionStats :one
SELECT * FROM champion_stats ORDER BY created_at DESC LIMIT 1;
