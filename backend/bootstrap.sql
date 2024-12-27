CREATE TABLE IF NOT EXISTS users (
    id text,
    nickname text,
    email text,
    google_user_id text,
    color_preferences text
);

CREATE TABLE IF NOT EXISTS games (
    id text,
    created_at int,
    name text,
    num_players int,
    map_name text,
    owner_user_id text,
    started int,
    finished int,
    game_state text,
    active_player_id text
);

CREATE TABLE IF NOT EXISTS game_player_map (
    game_id text,
    player_user_id text
);

CREATE TABLE IF NOT EXISTS game_log (
    game_id text,
    timestamp int, --Epoch seconds
    user_id text,
    action text,
    description text,
    new_active_player text,
    new_game_state text
);
