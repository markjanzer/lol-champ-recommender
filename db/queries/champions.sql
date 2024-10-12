-- name: CreateChampion :exec
INSERT INTO champions (api_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING;