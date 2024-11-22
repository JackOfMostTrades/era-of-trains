CREATE TABLE IF NOT EXISTS users (
    id text,
    nickname text,
    email text,
    google_user_id text
);

CREATE TABLE IF NOT EXISTS games (
    id text,
    num_players int,
    map_name text,
    owner_user_id text,
    started int,
    game_state text
);

CREATE TABLE IF NOT EXISTS game_player_map (
    game_id text,
    player_user_id text
);
