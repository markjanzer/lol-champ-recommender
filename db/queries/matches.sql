-- name: CreateMatch :exec
INSERT INTO matches (
    match_id, 
    game_start, 
    game_version, 
    winning_team, 
    queue_id,
    server_id,
    blue_1_champion_id, 
    blue_2_champion_id, 
    blue_3_champion_id, 
    blue_4_champion_id, 
    blue_5_champion_id, 
    red_1_champion_id, 
    red_2_champion_id, 
    red_3_champion_id, 
    red_4_champion_id, 
    red_5_champion_id
  ) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id;

-- name: AllMatchIds :many
SELECT matches.id FROM matches;

-- name: LastMatches :many
SELECT matches.match_id FROM matches ORDER BY created_at DESC LIMIT 10;

-- name: LastMatchesFromServer :many
SELECT matches.match_id FROM matches WHERE server_id = $1 ORDER BY created_at DESC LIMIT 10;

-- name: LastMatch :one
SELECT * FROM matches ORDER BY created_at DESC LIMIT 1;

-- name: AnyMatchesFromServer :one
SELECT EXISTS(SELECT 1 FROM matches WHERE server_id = $1);

-- name: AnyMatches :one
SELECT EXISTS(SELECT 1 FROM matches);

-- name: MatchExists :one
SELECT EXISTS(SELECT 1 FROM matches WHERE match_id = $1);

-- name: GetMatch :one
SELECT * FROM matches WHERE id = $1;