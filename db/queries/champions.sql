-- name: CreateChampion :exec
INSERT INTO champions (api_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetChampionsNotIn :many
SELECT * FROM champions WHERE id NOT IN ($1);

-- name: AllChampionIds :many
SELECT id FROM champions;