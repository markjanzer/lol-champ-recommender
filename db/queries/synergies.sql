-- name: CreateOrUpdateSynergy :exec
INSERT INTO synergies (champion1_id, champion2_id, wins, games_played) VALUES ($1, $2, $3, 1) ON CONFLICT (champion1_id, champion2_id) DO UPDATE SET wins = synergies.wins + $3, games_played = synergies.games_played + 1 RETURNING *;
