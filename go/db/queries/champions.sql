-- name: UpsertChampion :exec
INSERT INTO champions (api_id, name)
VALUES ($1, $2)
ON CONFLICT (api_id) DO UPDATE SET name = EXCLUDED.name;

-- name: GetChampionsNotIn :many
SELECT * FROM champions WHERE id NOT IN ($1);

-- name: AllChampionIds :many
SELECT id FROM champions;

-- name: AllChampionRiotIds :many 
SELECT api_id FROM champions;

-- name: AllChampions :many
SELECT api_id, name FROM champions;