BEGIN;
ALTER TABLE flight_states
    ADD COLUMN message_observed_at timestamptz;
COMMENT ON COLUMN flight_states.observed_at IS
    'Time of the position observation used by canonical traffic, trajectory, freshness, and analytics semantics.';
COMMENT ON COLUMN flight_states.message_observed_at IS
    'Optional time of the latest provider message/contact; may be newer than the position observation and must not replace position freshness.';
COMMIT;
