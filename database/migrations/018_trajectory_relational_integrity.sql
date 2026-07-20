BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM trajectory_segments
        WHERE trajectory_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'trajectory_segments contains rows without trajectory_id';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM coverage_gaps
        WHERE trajectory_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'coverage_gaps contains rows without trajectory_id';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM trajectory_segments
        WHERE sequence_number <= 0
    ) THEN
        RAISE EXCEPTION
            'trajectory_segments contains non-positive sequence_number values';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM trajectory_segments
        GROUP BY trajectory_id, sequence_number
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION
            'trajectory_segments contains duplicate trajectory sequence numbers';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM coverage_gaps AS gap
        LEFT JOIN trajectory_segments AS previous_segment
            ON previous_segment.id = gap.previous_segment_id
        WHERE gap.previous_segment_id IS NOT NULL
            AND (
                previous_segment.id IS NULL
                OR previous_segment.trajectory_id <> gap.trajectory_id
            )
    ) THEN
        RAISE EXCEPTION
            'coverage_gaps contains previous_segment_id from another trajectory';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM coverage_gaps AS gap
        LEFT JOIN trajectory_segments AS next_segment
            ON next_segment.id = gap.next_segment_id
        WHERE gap.next_segment_id IS NOT NULL
            AND (
                next_segment.id IS NULL
                OR next_segment.trajectory_id <> gap.trajectory_id
            )
    ) THEN
        RAISE EXCEPTION
            'coverage_gaps contains next_segment_id from another trajectory';
    END IF;

    IF EXISTS (
        WITH segment_stats AS (
            SELECT
                trajectory_id,
                COUNT(*)::integer AS actual_segment_count,
                COALESCE(SUM(point_count), 0)::integer AS actual_point_count,
                MIN(sequence_number) AS minimum_sequence_number,
                MAX(sequence_number) AS maximum_sequence_number,
                BOOL_AND(segment.icao24 = parent.icao24) AS identity_matches
            FROM trajectory_segments AS segment
            JOIN flight_trajectories AS parent
                ON parent.id = segment.trajectory_id
            GROUP BY trajectory_id
        ),
        gap_stats AS (
            SELECT
                gap.trajectory_id,
                COUNT(*)::integer AS actual_gap_count,
                BOOL_AND(gap.icao24 = parent.icao24) AS identity_matches
            FROM coverage_gaps AS gap
            JOIN flight_trajectories AS parent
                ON parent.id = gap.trajectory_id
            GROUP BY gap.trajectory_id
        )
        SELECT 1
        FROM flight_trajectories AS trajectory
        LEFT JOIN segment_stats AS segments
            ON segments.trajectory_id = trajectory.id
        LEFT JOIN gap_stats AS gaps
            ON gaps.trajectory_id = trajectory.id
        WHERE trajectory.segment_count < 0
            OR trajectory.point_count < 0
            OR trajectory.coverage_gap_count < 0
            OR trajectory.segment_count <> COALESCE(segments.actual_segment_count, 0)
            OR trajectory.point_count <> COALESCE(segments.actual_point_count, 0)
            OR trajectory.coverage_gap_count <> COALESCE(gaps.actual_gap_count, 0)
            OR COALESCE(segments.identity_matches, TRUE) = FALSE
            OR COALESCE(gaps.identity_matches, TRUE) = FALSE
            OR (
                COALESCE(segments.actual_segment_count, 0) > 0
                AND (
                    segments.minimum_sequence_number <> 1
                    OR segments.maximum_sequence_number <> segments.actual_segment_count
                )
            )
    ) THEN
        RAISE EXCEPTION
            'flight trajectory stored counts or child identity are inconsistent';
    END IF;
END;
$$;

ALTER TABLE trajectory_segments
    ALTER COLUMN trajectory_id SET NOT NULL,
    ADD CONSTRAINT trajectory_segments_sequence_number_positive_check
        CHECK (sequence_number > 0),
    ADD CONSTRAINT trajectory_segments_trajectory_sequence_unique
        UNIQUE (trajectory_id, sequence_number),
    ADD CONSTRAINT trajectory_segments_trajectory_id_id_unique
        UNIQUE (trajectory_id, id);

ALTER TABLE coverage_gaps
    ALTER COLUMN trajectory_id SET NOT NULL,
    ADD CONSTRAINT coverage_gaps_distinct_segment_references_check
        CHECK (
            previous_segment_id IS NULL
            OR next_segment_id IS NULL
            OR previous_segment_id <> next_segment_id
        ),
    ADD CONSTRAINT coverage_gaps_previous_segment_same_trajectory_fk
        FOREIGN KEY (trajectory_id, previous_segment_id)
        REFERENCES trajectory_segments (trajectory_id, id)
        DEFERRABLE INITIALLY DEFERRED,
    ADD CONSTRAINT coverage_gaps_next_segment_same_trajectory_fk
        FOREIGN KEY (trajectory_id, next_segment_id)
        REFERENCES trajectory_segments (trajectory_id, id)
        DEFERRABLE INITIALLY DEFERRED;

CREATE OR REPLACE FUNCTION assert_flight_trajectory_relational_integrity(
    target_trajectory_id uuid
)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    parent_record flight_trajectories%ROWTYPE;
    actual_segment_count integer;
    actual_point_count integer;
    minimum_sequence_number integer;
    maximum_sequence_number integer;
    segment_identity_matches boolean;
    actual_gap_count integer;
    gap_identity_matches boolean;
BEGIN
    SELECT *
    INTO parent_record
    FROM flight_trajectories
    WHERE id = target_trajectory_id;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    SELECT
        COUNT(*)::integer,
        COALESCE(SUM(point_count), 0)::integer,
        MIN(sequence_number),
        MAX(sequence_number),
        COALESCE(BOOL_AND(icao24 = parent_record.icao24), TRUE)
    INTO
        actual_segment_count,
        actual_point_count,
        minimum_sequence_number,
        maximum_sequence_number,
        segment_identity_matches
    FROM trajectory_segments
    WHERE trajectory_id = target_trajectory_id;

    SELECT
        COUNT(*)::integer,
        COALESCE(BOOL_AND(icao24 = parent_record.icao24), TRUE)
    INTO
        actual_gap_count,
        gap_identity_matches
    FROM coverage_gaps
    WHERE trajectory_id = target_trajectory_id;

    IF parent_record.segment_count <> actual_segment_count
        OR parent_record.point_count <> actual_point_count
        OR parent_record.coverage_gap_count <> actual_gap_count
        OR segment_identity_matches = FALSE
        OR gap_identity_matches = FALSE
        OR (
            actual_segment_count > 0
            AND (
                minimum_sequence_number <> 1
                OR maximum_sequence_number <> actual_segment_count
            )
        )
    THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = format(
                'flight trajectory %s relational integrity violation: stored segments=%s actual segments=%s stored points=%s actual points=%s stored gaps=%s actual gaps=%s',
                target_trajectory_id,
                parent_record.segment_count,
                actual_segment_count,
                parent_record.point_count,
                actual_point_count,
                parent_record.coverage_gap_count,
                actual_gap_count
            );
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION enforce_flight_trajectory_relational_integrity()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_TABLE_NAME = 'flight_trajectories' THEN
        PERFORM assert_flight_trajectory_relational_integrity(NEW.id);
        RETURN NULL;
    END IF;

    IF TG_OP = 'DELETE' THEN
        PERFORM assert_flight_trajectory_relational_integrity(OLD.trajectory_id);
        RETURN NULL;
    END IF;

    IF TG_OP = 'UPDATE' AND OLD.trajectory_id IS DISTINCT FROM NEW.trajectory_id THEN
        PERFORM assert_flight_trajectory_relational_integrity(OLD.trajectory_id);
    END IF;

    PERFORM assert_flight_trajectory_relational_integrity(NEW.trajectory_id);
    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER flight_trajectories_relational_integrity_trigger
AFTER INSERT OR UPDATE ON flight_trajectories
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION enforce_flight_trajectory_relational_integrity();

CREATE CONSTRAINT TRIGGER trajectory_segments_relational_integrity_trigger
AFTER INSERT OR UPDATE OR DELETE ON trajectory_segments
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION enforce_flight_trajectory_relational_integrity();

CREATE CONSTRAINT TRIGGER coverage_gaps_relational_integrity_trigger
AFTER INSERT OR UPDATE OR DELETE ON coverage_gaps
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION enforce_flight_trajectory_relational_integrity();

COMMIT;
