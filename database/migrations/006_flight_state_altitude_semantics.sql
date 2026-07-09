BEGIN;

ALTER TABLE flight_states
    ADD COLUMN barometric_altitude_status text,
    ADD COLUMN geometric_altitude_status text;

UPDATE flight_states
SET
    barometric_altitude_status = CASE
        WHEN barometric_altitude_m IS NULL THEN 'unavailable'
        WHEN barometric_altitude_m = 0 AND on_ground IS TRUE THEN 'ground'
        WHEN barometric_altitude_m = 0 THEN 'unknown'
        ELSE 'observed'
    END,
    geometric_altitude_status = CASE
        WHEN geometric_altitude_m IS NULL THEN 'unavailable'
        WHEN geometric_altitude_m = 0 THEN 'unknown'
        ELSE 'observed'
    END;

-- Legacy zero values are semantically ambiguous because the old application
-- collapsed missing, unknown, invalid, ground, and numeric zero into the same
-- numeric representation. Preserve only evidence that can be reconstructed.
UPDATE flight_states
SET barometric_altitude_m = NULL
WHERE barometric_altitude_status = 'unknown';

UPDATE flight_states
SET geometric_altitude_m = NULL
WHERE geometric_altitude_status = 'unknown';

ALTER TABLE flight_states
    ALTER COLUMN barometric_altitude_status SET NOT NULL,
    ALTER COLUMN geometric_altitude_status SET NOT NULL;

ALTER TABLE flight_states
    ADD CONSTRAINT flight_states_barometric_altitude_status_check
        CHECK (
            barometric_altitude_status IN (
                'observed',
                'ground',
                'unknown',
                'unavailable',
                'invalid'
            )
        ),
    ADD CONSTRAINT flight_states_geometric_altitude_status_check
        CHECK (
            geometric_altitude_status IN (
                'observed',
                'ground',
                'unknown',
                'unavailable',
                'invalid'
            )
        ),
    ADD CONSTRAINT flight_states_barometric_altitude_semantics_check
        CHECK (
            (
                barometric_altitude_status = 'observed'
                AND barometric_altitude_m IS NOT NULL
            )
            OR (
                barometric_altitude_status = 'ground'
                AND barometric_altitude_m = 0
                AND on_ground IS TRUE
            )
            OR (
                barometric_altitude_status IN (
                    'unknown',
                    'unavailable',
                    'invalid'
                )
                AND barometric_altitude_m IS NULL
            )
        ),
    ADD CONSTRAINT flight_states_geometric_altitude_semantics_check
        CHECK (
            (
                geometric_altitude_status = 'observed'
                AND geometric_altitude_m IS NOT NULL
            )
            OR (
                geometric_altitude_status = 'ground'
                AND geometric_altitude_m = 0
                AND on_ground IS TRUE
            )
            OR (
                geometric_altitude_status IN (
                    'unknown',
                    'unavailable',
                    'invalid'
                )
                AND geometric_altitude_m IS NULL
            )
        );

COMMIT;
