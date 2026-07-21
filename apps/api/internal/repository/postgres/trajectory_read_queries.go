package postgres

const flightTrajectorySelectColumns = `
	trajectory.id::text,
	COALESCE(trajectory.identity_key, ''),
	COALESCE(trajectory.identity_basis, ''),
	COALESCE(trajectory.split_reason, ''),
	COALESCE(trajectory.flight_id::text, ''),
	COALESCE(trajectory.aircraft_id::text, ''),
	trajectory.icao24,
	COALESCE(trajectory.callsign, ''),
	trajectory.start_time,
	trajectory.end_time,
	trajectory.duration_seconds::bigint,
	trajectory.segment_count,
	trajectory.point_count,
	trajectory.coverage_gap_count,
	trajectory.quality_score::float8,
	trajectory.source_name,
	trajectory.created_at,
	trajectory.updated_at
`

const latestTrajectoryByICAO24Query = `
	SELECT ` + flightTrajectorySelectColumns + `
	FROM flight_trajectories AS trajectory
	WHERE trajectory.icao24 = $1
	ORDER BY
		trajectory.end_time DESC,
		trajectory.start_time DESC,
		trajectory.created_at DESC
	LIMIT 1;
`

const trajectoryByIDQuery = `
	SELECT ` + flightTrajectorySelectColumns + `
	FROM flight_trajectories AS trajectory
	WHERE trajectory.id = $1
	LIMIT 1;
`

const trajectoriesByEndTimeQuery = `
	SELECT ` + flightTrajectorySelectColumns + `
	FROM flight_trajectories AS trajectory
	WHERE trajectory.end_time >= $1
		AND trajectory.end_time <= $2
	ORDER BY
		trajectory.end_time DESC,
		trajectory.start_time DESC,
		trajectory.created_at DESC
	LIMIT $3;
`

const trajectoriesByIDsQuery = `
	SELECT ` + flightTrajectorySelectColumns + `
	FROM unnest($1::text[]) WITH ORDINALITY AS requested(id_text, requested_position)
	INNER JOIN flight_trajectories AS trajectory
		ON trajectory.id = requested.id_text::uuid
	ORDER BY requested.requested_position;
`

const trajectoriesByEndTimeAndBoundsQuery = `
	SELECT ` + flightTrajectorySelectColumns + `
	FROM flight_trajectories AS trajectory
	JOIN LATERAL (
		SELECT
			segment.end_latitude::float8 AS latitude,
			segment.end_longitude::float8 AS longitude
		FROM trajectory_segments AS segment
		WHERE segment.trajectory_id = trajectory.id
			AND segment.status <> 'invalid'
		ORDER BY
			segment.sequence_number DESC,
			segment.end_time DESC,
			segment.created_at DESC
		LIMIT 1
	) AS latest_position ON TRUE
	WHERE trajectory.end_time >= $1
		AND trajectory.end_time <= $2
		AND latest_position.latitude BETWEEN $3 AND $4
		AND latest_position.longitude BETWEEN $5 AND $6
	ORDER BY
		trajectory.end_time DESC,
		trajectory.start_time DESC,
		trajectory.created_at DESC
	LIMIT $7;
`

const trajectorySegmentSelectColumns = `
	segment.id::text,
	segment.trajectory_id::text,
	COALESCE(segment.flight_id::text, ''),
	COALESCE(segment.aircraft_id::text, ''),
	segment.icao24,
	COALESCE(segment.callsign, ''),
	segment.sequence_number,
	segment.status,
	segment.quality_score::float8,
	segment.start_time,
	segment.end_time,
	segment.duration_seconds::bigint,
	segment.start_latitude::float8,
	segment.start_longitude::float8,
	segment.end_latitude::float8,
	segment.end_longitude::float8,
	segment.point_count,
	segment.source_name,
	segment.created_at
`

const trajectorySegmentsByTrajectoryIDQuery = `
	SELECT ` + trajectorySegmentSelectColumns + `
	FROM trajectory_segments AS segment
	WHERE segment.trajectory_id = $1
	ORDER BY segment.sequence_number ASC;
`

const coverageGapSelectColumns = `
	gap.id::text,
	gap.trajectory_id::text,
	COALESCE(gap.previous_segment_id::text, ''),
	COALESCE(gap.next_segment_id::text, ''),
	gap.icao24,
	gap.gap_start_time,
	gap.gap_end_time,
	gap.duration_seconds::bigint,
	gap.distance_km::float8,
	gap.reason,
	COALESCE(gap.filled_by, ''),
	gap.created_at
`

const coverageGapsByTrajectoryIDQuery = `
	SELECT ` + coverageGapSelectColumns + `
	FROM coverage_gaps AS gap
	WHERE gap.trajectory_id = $1
	ORDER BY gap.gap_start_time ASC;
`
