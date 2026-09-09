BEGIN;

ALTER TABLE flight_states
    DROP CONSTRAINT flight_states_position_source_check;

ALTER TABLE flight_states
    ADD CONSTRAINT flight_states_position_source_check
        CHECK (
            position_source IN (
                '',
                'adsb',
                'adsr',
                'tisb',
                'adsc',
                'asterix',
                'mlat',
                'flarm'
            )
        );

COMMENT ON COLUMN flight_states.position_source IS
    'Canonical observed position method. Empty means unavailable; provider/feed identity remains in source_name.';

COMMIT;
