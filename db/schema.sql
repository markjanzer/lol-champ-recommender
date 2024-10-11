CREATE TABLE matches (
  id SERIAL PRIMARY KEY,
  match_id VARCHAR(255) NOT NULL,
  region VARCHAR(255) NOT NULL,
  game_start TIMESTAMP NOT NULL,
  game_version VARCHAR(255) NOT NULL,
  winning_team VARCHAR(255) NOT NULL,
  red_1_champion_id INTEGER NOT NULL,
  red_2_champion_id INTEGER NOT NULL,
  red_3_champion_id INTEGER NOT NULL,
  red_4_champion_id INTEGER NOT NULL,
  red_5_champion_id INTEGER NOT NULL,
  blue_1_champion_id INTEGER NOT NULL,
  blue_2_champion_id INTEGER NOT NULL,
  blue_3_champion_id INTEGER NOT NULL,
  blue_4_champion_id INTEGER NOT NULL,
  blue_5_champion_id INTEGER NOT NULL
)

CREATE TABLE champions (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  api_id INTEGER NOT NULL
)