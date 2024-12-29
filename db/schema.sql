CREATE TABLE IF NOT EXISTS matches (
  id SERIAL PRIMARY KEY,
  match_id VARCHAR(255) NOT NULL UNIQUE,
  game_start TIMESTAMP NOT NULL,
  game_version VARCHAR(255) NOT NULL,
  winning_team VARCHAR(255) NOT NULL,
  queue_id INTEGER NOT NULL,
  red_1_champion_id INTEGER NOT NULL,
  red_2_champion_id INTEGER NOT NULL,
  red_3_champion_id INTEGER NOT NULL,
  red_4_champion_id INTEGER NOT NULL,
  red_5_champion_id INTEGER NOT NULL,
  blue_1_champion_id INTEGER NOT NULL,
  blue_2_champion_id INTEGER NOT NULL,
  blue_3_champion_id INTEGER NOT NULL,
  blue_4_champion_id INTEGER NOT NULL,
  blue_5_champion_id INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS champions (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  api_id INTEGER UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS player_search_log (
  id SERIAL PRIMARY KEY,
  player_id VARCHAR(255) NOT NULL CHECK (player_id <> ''),
  search_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS champion_stats (
  id SERIAL PRIMARY KEY,
  data JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_match_id ON matches(match_id);