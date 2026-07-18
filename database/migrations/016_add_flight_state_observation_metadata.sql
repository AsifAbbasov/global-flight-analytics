BEGIN;

ALTER TABLE flight_states
    ADD COLUMN squawk_code varchar(4) NOT NULL DEFAULT '',
    ADD COLUMN special_purpose_indicator boolean NOT NULL DEFAULT false,
    ADD COLUMN position_source text NOT NULL DEFAULT '',
    ADD COLUMN aircraft_category smallint,
    ADD COLUMN aircraft_category_available boolean NOT NULL DEFAULT false;

ALTER TABLE flight_states
    ADD CONSTRAINT flight_states_squawk_code_check
        CHECK (
            squawk_code = ''
            OR squawk_code ~ '^[0-7]{4}$'
        ),
    ADD CONSTRAINT flight_states_position_source_check
        CHECK (
            position_source IN (
                '',
                'adsb',
                'asterix',
                'mlat',
                'flarm'
            )
        ),
    ADD CONSTRAINT flight_states_aircraft_category_check
        CHECK (
            (
                aircraft_category_available = false
                AND aircraft_category IS NULL
            )
            OR
            (
                aircraft_category_available = true
                AND aircraft_category BETWEEN 0 AND 20
            )
        );

CREATE INDEX flight_states_special_transponder_code_idx
    ON flight_states (
        squawk_code,
        observed_at DESC,
        icao24
    )
    WHERE squawk_code IN (
        '7500',
        '7600',
        '7700'
    );

CREATE INDEX flight_states_position_source_observed_at_idx
    ON flight_states (
        position_source,
        observed_at DESC
    )
    WHERE position_source <> '';

COMMIT;
