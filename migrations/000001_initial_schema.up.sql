CREATE TABLE IF NOT EXISTS players (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR,
    phone VARCHAR,
    email VARCHAR,
    username VARCHAR,
    password_digest VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR,
    street VARCHAR,
    city VARCHAR,
    state VARCHAR,
    zip_code VARCHAR,
    phone VARCHAR,
    cost VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    course_id INTEGER,
    date VARCHAR,
    tee_time VARCHAR,
    open_spots INTEGER,
    number_of_holes VARCHAR,
    private BOOLEAN,
    host_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS friendships (
    id BIGSERIAL PRIMARY KEY,
    follower_id INTEGER,
    followee_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS player_events (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT,
    event_id BIGINT,
    invite_status INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS index_player_events_on_event_id ON player_events (event_id);
CREATE INDEX IF NOT EXISTS index_player_events_on_player_id ON player_events (player_id);

ALTER TABLE player_events
    ADD CONSTRAINT fk_player_events_players FOREIGN KEY (player_id) REFERENCES players(id);
ALTER TABLE player_events
    ADD CONSTRAINT fk_player_events_events FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
