-- Reverse 000010_reservations.

DROP INDEX IF EXISTS idx_reservation_players_reservation_id;
DROP TABLE IF EXISTS reservation_players;

DROP INDEX IF EXISTS idx_reservations_external;
DROP INDEX IF EXISTS idx_reservations_booked_by;
DROP INDEX IF EXISTS uq_reservations_active_tee_time;
DROP TABLE IF EXISTS reservations;
