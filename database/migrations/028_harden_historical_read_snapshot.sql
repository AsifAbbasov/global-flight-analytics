BEGIN;

CREATE TABLE historical_read_history_state (
    singleton boolean PRIMARY KEY DEFAULT true,
    coverage_started_at timestamptz NOT NULL,

    CONSTRAINT historical_read_history_state_singleton_check
        CHECK (singleton)
);

INSERT INTO historical_read_history_state (
    singleton,
    coverage_started_at
)
VALUES (
    true,
    clock_timestamp()
);

CREATE TABLE historical_read_flight_versions (
    version_id bigserial PRIMARY KEY,
    flight_id uuid NOT NULL,
    aircraft_id uuid,
    callsign text,
    status text NOT NULL,
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    source_created_at timestamptz NOT NULL,
    source_updated_at timestamptz NOT NULL,
    recorded_from timestamptz NOT NULL,
    recorded_to timestamptz,

    CONSTRAINT historical_read_flight_versions_window_check
        CHECK (first_seen_at <= last_seen_at),

    CONSTRAINT historical_read_flight_versions_recorded_check
        CHECK (recorded_to IS NULL OR recorded_from <= recorded_to)
);

CREATE TABLE historical_read_trajectory_versions (
    version_id bigserial PRIMARY KEY,
    trajectory_id uuid NOT NULL,
    flight_id uuid,
    aircraft_id uuid,
    icao24 varchar(10) NOT NULL,
    callsign text,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    segment_count integer NOT NULL,
    point_count integer NOT NULL,
    coverage_gap_count integer NOT NULL,
    quality_score numeric NOT NULL,
    source_name text NOT NULL,
    source_created_at timestamptz NOT NULL,
    source_updated_at timestamptz NOT NULL,
    recorded_from timestamptz NOT NULL,
    recorded_to timestamptz,

    CONSTRAINT historical_read_trajectory_versions_window_check
        CHECK (start_time <= end_time),

    CONSTRAINT historical_read_trajectory_versions_counts_check
        CHECK (
            segment_count >= 0
            AND point_count >= 0
            AND coverage_gap_count >= 0
        ),

    CONSTRAINT historical_read_trajectory_versions_quality_check
        CHECK (quality_score >= 0 AND quality_score <= 1),

    CONSTRAINT historical_read_trajectory_versions_recorded_check
        CHECK (recorded_to IS NULL OR recorded_from <= recorded_to)
);

INSERT INTO historical_read_flight_versions (
    flight_id,
    aircraft_id,
    callsign,
    status,
    first_seen_at,
    last_seen_at,
    source_created_at,
    source_updated_at,
    recorded_from,
    recorded_to
)
SELECT
    id,
    aircraft_id,
    callsign,
    status,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at,
    updated_at,
    NULL
FROM flights;

INSERT INTO historical_read_trajectory_versions (
    trajectory_id,
    flight_id,
    aircraft_id,
    icao24,
    callsign,
    start_time,
    end_time,
    segment_count,
    point_count,
    coverage_gap_count,
    quality_score,
    source_name,
    source_created_at,
    source_updated_at,
    recorded_from,
    recorded_to
)
SELECT
    id,
    flight_id,
    aircraft_id,
    icao24,
    callsign,
    start_time,
    end_time,
    segment_count,
    point_count,
    coverage_gap_count,
    quality_score,
    source_name,
    created_at,
    updated_at,
    updated_at,
    NULL
FROM flight_trajectories;

CREATE OR REPLACE FUNCTION capture_historical_read_flight_version()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    change_time timestamptz;
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO historical_read_flight_versions (
            flight_id,
            aircraft_id,
            callsign,
            status,
            first_seen_at,
            last_seen_at,
            source_created_at,
            source_updated_at,
            recorded_from,
            recorded_to
        )
        VALUES (
            NEW.id,
            NEW.aircraft_id,
            NEW.callsign,
            NEW.status,
            NEW.first_seen_at,
            NEW.last_seen_at,
            NEW.created_at,
            NEW.updated_at,
            NEW.updated_at,
            NULL
        );
        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        change_time := GREATEST(NEW.updated_at, OLD.updated_at);

        UPDATE historical_read_flight_versions
        SET recorded_to = change_time
        WHERE flight_id = OLD.id
          AND recorded_to IS NULL;

        INSERT INTO historical_read_flight_versions (
            flight_id,
            aircraft_id,
            callsign,
            status,
            first_seen_at,
            last_seen_at,
            source_created_at,
            source_updated_at,
            recorded_from,
            recorded_to
        )
        VALUES (
            NEW.id,
            NEW.aircraft_id,
            NEW.callsign,
            NEW.status,
            NEW.first_seen_at,
            NEW.last_seen_at,
            NEW.created_at,
            NEW.updated_at,
            change_time,
            NULL
        );
        RETURN NEW;
    END IF;

    UPDATE historical_read_flight_versions
    SET recorded_to = clock_timestamp()
    WHERE flight_id = OLD.id
      AND recorded_to IS NULL;

    RETURN OLD;
END;
$$;

CREATE OR REPLACE FUNCTION capture_historical_read_trajectory_version()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    change_time timestamptz;
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO historical_read_trajectory_versions (
            trajectory_id,
            flight_id,
            aircraft_id,
            icao24,
            callsign,
            start_time,
            end_time,
            segment_count,
            point_count,
            coverage_gap_count,
            quality_score,
            source_name,
            source_created_at,
            source_updated_at,
            recorded_from,
            recorded_to
        )
        VALUES (
            NEW.id,
            NEW.flight_id,
            NEW.aircraft_id,
            NEW.icao24,
            NEW.callsign,
            NEW.start_time,
            NEW.end_time,
            NEW.segment_count,
            NEW.point_count,
            NEW.coverage_gap_count,
            NEW.quality_score,
            NEW.source_name,
            NEW.created_at,
            NEW.updated_at,
            NEW.updated_at,
            NULL
        );
        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        change_time := GREATEST(NEW.updated_at, OLD.updated_at);

        UPDATE historical_read_trajectory_versions
        SET recorded_to = change_time
        WHERE trajectory_id = OLD.id
          AND recorded_to IS NULL;

        INSERT INTO historical_read_trajectory_versions (
            trajectory_id,
            flight_id,
            aircraft_id,
            icao24,
            callsign,
            start_time,
            end_time,
            segment_count,
            point_count,
            coverage_gap_count,
            quality_score,
            source_name,
            source_created_at,
            source_updated_at,
            recorded_from,
            recorded_to
        )
        VALUES (
            NEW.id,
            NEW.flight_id,
            NEW.aircraft_id,
            NEW.icao24,
            NEW.callsign,
            NEW.start_time,
            NEW.end_time,
            NEW.segment_count,
            NEW.point_count,
            NEW.coverage_gap_count,
            NEW.quality_score,
            NEW.source_name,
            NEW.created_at,
            NEW.updated_at,
            change_time,
            NULL
        );
        RETURN NEW;
    END IF;

    UPDATE historical_read_trajectory_versions
    SET recorded_to = clock_timestamp()
    WHERE trajectory_id = OLD.id
      AND recorded_to IS NULL;

    RETURN OLD;
END;
$$;

CREATE TRIGGER flights_historical_read_version_trigger
AFTER INSERT OR UPDATE OR DELETE ON flights
FOR EACH ROW
EXECUTE FUNCTION capture_historical_read_flight_version();

CREATE TRIGGER flight_trajectories_historical_read_version_trigger
AFTER INSERT OR UPDATE OR DELETE ON flight_trajectories
FOR EACH ROW
EXECUTE FUNCTION capture_historical_read_trajectory_version();

CREATE INDEX historical_read_flight_versions_event_idx
    ON historical_read_flight_versions (
        first_seen_at ASC,
        flight_id ASC
    )
    INCLUDE (
        last_seen_at,
        recorded_from,
        recorded_to,
        source_updated_at
    );

CREATE INDEX historical_read_flight_versions_identity_idx
    ON historical_read_flight_versions (
        flight_id,
        recorded_from DESC,
        version_id DESC
    );

CREATE UNIQUE INDEX historical_read_flight_versions_current_idx
    ON historical_read_flight_versions (flight_id)
    WHERE recorded_to IS NULL;

CREATE INDEX historical_read_trajectory_versions_event_idx
    ON historical_read_trajectory_versions (
        start_time ASC,
        trajectory_id ASC
    )
    INCLUDE (
        end_time,
        recorded_from,
        recorded_to,
        source_updated_at
    );

CREATE INDEX historical_read_trajectory_versions_identity_idx
    ON historical_read_trajectory_versions (
        trajectory_id,
        recorded_from DESC,
        version_id DESC
    );

CREATE UNIQUE INDEX historical_read_trajectory_versions_current_idx
    ON historical_read_trajectory_versions (trajectory_id)
    WHERE recorded_to IS NULL;

CREATE INDEX flight_states_historical_read_idx
    ON flight_states (
        observed_at ASC,
        id ASC
    )
    INCLUDE (
        created_at,
        flight_id,
        aircraft_id,
        icao24,
        callsign,
        latitude,
        longitude,
        on_ground,
        source_name
    );

CREATE INDEX flight_route_results_historical_read_idx
    ON flight_route_results (
        trajectory_id,
        as_of_time DESC,
        stored_at DESC,
        id ASC
    )
    INCLUDE (
        input_fingerprint,
        route_status,
        confidence_level,
        validation_warning_count
    );

COMMIT;
