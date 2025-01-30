CREATE TABLE IF NOT EXISTS users (
    id text PRIMARY KEY,
    nickname text,
    email text,
    google_user_id text,
    discord_user_id text,
    color_preferences text,
    custom_colors text,
    email_notifications_enabled int,
    discord_turn_alerts_enabled int,
    webhooks text
);

CREATE TABLE IF NOT EXISTS games (
    id text PRIMARY KEY,
    created_at int,
    name text,
    min_players int,
    max_players int,
    map_name text,
    owner_user_id text,
    started int,
    finished int,
    game_state text,
    active_player_id text,
    invite_only int
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
    new_game_state text,
    PRIMARY KEY (game_id, timestamp)
);

CREATE TABLE IF NOT EXISTS game_chat (
    game_id text,
    timestamp int, --Epoch seconds
    user_id text,
    message text,
    PRIMARY KEY (game_id, timestamp)
);
