BEGIN;

-- Migration 018 introduced a unique index for the same key. Retire the older
-- non-unique duplicate so every write maintains only one segment-order index.
DROP INDEX trajectory_segments_trajectory_sequence_idx;

CREATE INDEX flight_trajectories_icao24_latest_idx
    ON flight_trajectories (
        icao24,
        end_time DESC,
        start_time DESC,
        created_at DESC
    );

CREATE INDEX flight_trajectories_end_time_order_idx
    ON flight_trajectories (
        end_time DESC,
        start_time DESC,
        created_at DESC
    );

COMMIT;
