BEGIN;

CREATE TABLE flight_trajectories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id uuid REFERENCES flights(id) ON DELETE SET NULL,
    aircraft_id uuid REFERENCES aircraft(id) ON DELETE SET NULL,
    icao24 varchar(10) NOT NULL,
    callsign text,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    duration_seconds integer NOT NULL DEFAULT 0,
    segment_count integer NOT NULL DEFAULT 0,
    point_count integer NOT NULL DEFAULT 0,
    coverage_gap_count integer NOT NULL DEFAULT 0,
    quality_score numeric NOT NULL DEFAULT 0,
    source_name text NOT NULL,
    metadata_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT flight_trajectories_time_check
        CHECK (start_time <= end_time),

    CONSTRAINT flight_trajectories_counts_check
        CHECK (
            duration_seconds >= 0
            AND segment_count >= 0
            AND point_count >= 0
            AND coverage_gap_count >= 0
        ),

    CONSTRAINT flight_trajectories_quality_score_check
        CHECK (quality_score >= 0 AND quality_score <= 1)
);

CREATE TABLE trajectory_segments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trajectory_id uuid REFERENCES flight_trajectories(id) ON DELETE CASCADE,
    flight_id uuid REFERENCES flights(id) ON DELETE SET NULL,
    aircraft_id uuid REFERENCES aircraft(id) ON DELETE SET NULL,
    icao24 varchar(10) NOT NULL,
    callsign text,
    sequence_number integer NOT NULL,
    status text NOT NULL,
    quality_score numeric NOT NULL DEFAULT 0,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    duration_seconds integer NOT NULL DEFAULT 0,
    start_latitude numeric NOT NULL,
    start_longitude numeric NOT NULL,
    end_latitude numeric NOT NULL,
    end_longitude numeric NOT NULL,
    point_count integer NOT NULL DEFAULT 0,
    source_name text NOT NULL,
    metadata_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT trajectory_segments_status_check
        CHECK (status IN ('observed', 'interpolated', 'estimated', 'invalid')),

    CONSTRAINT trajectory_segments_time_check
        CHECK (start_time <= end_time),

    CONSTRAINT trajectory_segments_counts_check
        CHECK (
            sequence_number > 0
            AND duration_seconds >= 0
            AND point_count >= 0
        ),

    CONSTRAINT trajectory_segments_quality_score_check
        CHECK (quality_score >= 0 AND quality_score <= 1),

    CONSTRAINT trajectory_segments_start_coordinates_check
        CHECK (
            start_latitude >= -90
            AND start_latitude <= 90
            AND start_longitude >= -180
            AND start_longitude <= 180
        ),

    CONSTRAINT trajectory_segments_end_coordinates_check
        CHECK (
            end_latitude >= -90
            AND end_latitude <= 90
            AND end_longitude >= -180
            AND end_longitude <= 180
        )
);

CREATE TABLE coverage_gaps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trajectory_id uuid REFERENCES flight_trajectories(id) ON DELETE CASCADE,
    previous_segment_id uuid REFERENCES trajectory_segments(id) ON DELETE SET NULL,
    next_segment_id uuid REFERENCES trajectory_segments(id) ON DELETE SET NULL,
    icao24 varchar(10) NOT NULL,
    gap_start_time timestamptz NOT NULL,
    gap_end_time timestamptz NOT NULL,
    duration_seconds integer NOT NULL DEFAULT 0,
    distance_km numeric NOT NULL DEFAULT 0,
    reason text NOT NULL,
    filled_by text,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT coverage_gaps_time_check
        CHECK (gap_start_time <= gap_end_time),

    CONSTRAINT coverage_gaps_values_check
        CHECK (
            duration_seconds >= 0
            AND distance_km >= 0
        ),

    CONSTRAINT coverage_gaps_reason_check
        CHECK (reason IN ('time_gap', 'movement_jump', 'unknown'))
);

CREATE TABLE data_quality_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_type text NOT NULL,
    object_id uuid,
    validation_status text NOT NULL,
    completeness text NOT NULL,
    confidence text NOT NULL,
    score numeric NOT NULL DEFAULT 0,
    missing_fields text[] NOT NULL DEFAULT '{}',
    warnings_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    calculated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT data_quality_reports_validation_status_check
        CHECK (validation_status IN ('valid', 'partial', 'invalid')),

    CONSTRAINT data_quality_reports_completeness_check
        CHECK (completeness IN ('complete', 'partial', 'position_only', 'insufficient')),

    CONSTRAINT data_quality_reports_confidence_check
        CHECK (confidence IN ('high', 'medium', 'low', 'none')),

    CONSTRAINT data_quality_reports_score_check
        CHECK (score >= 0 AND score <= 1)
);

CREATE INDEX flight_trajectories_icao24_time_idx
    ON flight_trajectories (icao24, start_time DESC, end_time DESC);

CREATE INDEX trajectory_segments_trajectory_sequence_idx
    ON trajectory_segments (trajectory_id, sequence_number);

CREATE INDEX trajectory_segments_icao24_time_idx
    ON trajectory_segments (icao24, start_time DESC, end_time DESC);

CREATE INDEX coverage_gaps_trajectory_time_idx
    ON coverage_gaps (trajectory_id, gap_start_time DESC);

CREATE INDEX data_quality_reports_object_idx
    ON data_quality_reports (object_type, object_id);

COMMIT;
