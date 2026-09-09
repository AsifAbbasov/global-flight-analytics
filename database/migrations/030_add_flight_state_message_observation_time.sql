BEGIN;
ALTER TABLE flight_states
    ADD COLUMN message_observed_at timestamptz;
COMMENT ON COLUMN flight_states.observed_at IS
    'Time of the position observation used by canonical traffic, trajectory, freshness, and analytics semantics.';
COMMENT ON COLUMN flight_states.message_observed_at IS
    'Optional latest provider message/contact time for this source, aircraft, and position identity; replay conflicts may monotonically refresh it without creating a new position state.';
COMMIT;
