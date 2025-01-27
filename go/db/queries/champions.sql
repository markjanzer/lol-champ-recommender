-- name: UpsertChampion :exec
INSERT INTO champions (api_id, name)
VALUES ($1, $2)
ON CONFLICT (api_id) DO UPDATE SET name = EXCLUDED.name;

-- name: AllChampionRiotIds :many 
SELECT api_id FROM champions;

-- name: AllChampions :many
SELECT api_id, name FROM champions;