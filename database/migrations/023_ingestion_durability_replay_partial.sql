BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM flight_states
        GROUP BY source_name, icao24, observed_at
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION
            'duplicate provider observation identities exist in flight_states; repair before migration 023';
    END IF;
END;
$$;

CREATE UNIQUE INDEX flight_states_source_observation_identity_idx
    ON flight_states (source_name, icao24, observed_at);

COMMIT;
