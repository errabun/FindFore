-- Party booking lifecycle. Event and reservation are siblings under a tee time.

CREATE TABLE IF NOT EXISTS reservations (
    id BIGSERIAL PRIMARY KEY,
    tee_time_id BIGINT NOT NULL REFERENCES tee_times(id),
    booked_by_player_id BIGINT NOT NULL REFERENCES players(id),
    status VARCHAR NOT NULL DEFAULT 'pending',
    party_size INTEGER NOT NULL,
    provider VARCHAR NOT NULL,
    external_reservation_id VARCHAR,
    hold_expires_at TIMESTAMPTZ,
    failure_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_reservations_status CHECK (
        status IN ('pending', 'held', 'confirmed', 'cancelled', 'failed')
    ),
    CONSTRAINT chk_reservations_party_size CHECK (party_size > 0)
);

-- At most one non-terminal reservation per tee time.
CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_active_tee_time
    ON reservations (tee_time_id)
    WHERE status IN ('pending', 'held', 'confirmed');

CREATE INDEX IF NOT EXISTS idx_reservations_booked_by
    ON reservations (booked_by_player_id);

CREATE INDEX IF NOT EXISTS idx_reservations_external
    ON reservations (provider, external_reservation_id)
    WHERE external_reservation_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS reservation_players (
    id BIGSERIAL PRIMARY KEY,
    reservation_id BIGINT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES players(id),
    guest_name VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_reservation_players_identity CHECK (
        player_id IS NOT NULL
        OR (guest_name IS NOT NULL AND BTRIM(guest_name) <> '')
    )
);

CREATE INDEX IF NOT EXISTS idx_reservation_players_reservation_id
    ON reservation_players (reservation_id);
