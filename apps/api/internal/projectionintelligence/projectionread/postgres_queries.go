package projectionread

const routeAtOrBeforeSQL = `
	SELECT route_json
	FROM flight_route_results
	WHERE trajectory_id = $1::uuid
	  AND schema_version = $2
	  AND as_of_time <= $3
	ORDER BY
		as_of_time DESC,
		id ASC
	LIMIT 1;
`

const historicalCandidateIDsSQL = `
	WITH latest_route_per_trajectory AS (
		SELECT DISTINCT ON (route_result.trajectory_id)
			route_result.trajectory_id,
			route_result.as_of_time
		FROM flight_route_results AS route_result
		WHERE route_result.schema_version = $1
		  AND route_result.route_status = 'complete'
		  AND route_result.as_of_time >= $2
		  AND route_result.as_of_time <= $3
		  AND route_result.trajectory_id <> $4::uuid
		  AND route_result.route_json #>>
				'{Origin,Airport,ICAOCode}' = $5
		  AND route_result.route_json #>>
				'{Destination,Airport,ICAOCode}' = $6
		ORDER BY
			route_result.trajectory_id,
			route_result.as_of_time DESC,
			route_result.id ASC
	)
	SELECT trajectory.id::text
	FROM latest_route_per_trajectory AS route_result
	INNER JOIN flight_trajectories AS trajectory
		ON trajectory.id = route_result.trajectory_id
	WHERE trajectory.end_time < $7
	ORDER BY
		trajectory.end_time DESC,
		trajectory.id ASC
	LIMIT $8;
`

const routeHistorySummarySQL = `
	WITH latest_route_per_trajectory AS (
		SELECT DISTINCT ON (route_result.trajectory_id)
			route_result.trajectory_id,
			route_result.as_of_time,
			trajectory.flight_id
		FROM flight_route_results AS route_result
		INNER JOIN flight_trajectories AS trajectory
			ON trajectory.id = route_result.trajectory_id
		WHERE route_result.schema_version = $1
		  AND route_result.route_status = 'complete'
		  AND route_result.as_of_time >= $2
		  AND route_result.as_of_time <= $3
		  AND route_result.route_json #>>
				'{Origin,Airport,ICAOCode}' = $4
		  AND route_result.route_json #>>
				'{Destination,Airport,ICAOCode}' = $5
		ORDER BY
			route_result.trajectory_id,
			route_result.as_of_time DESC,
			route_result.id ASC
	)
	SELECT
		COUNT(*)::bigint,
		COUNT(
			DISTINCT COALESCE(
				flight_id::text,
				trajectory_id::text
			)
		)::bigint,
		COUNT(
			DISTINCT (
				as_of_time AT TIME ZONE 'UTC'
			)::date
		)::bigint,
		COUNT(*) FILTER (
			WHERE as_of_time >= $6
		)::bigint,
		MAX(as_of_time)
	FROM latest_route_per_trajectory;
`

const trajectoryPointsByFlightSQL = `
	SELECT
		id::text,
		COALESCE(flight_id::text, ''),
		COALESCE(aircraft_id::text, ''),
		icao24,
		COALESCE(callsign, ''),
		latitude::float8,
		longitude::float8,
		barometric_altitude_m::float8,
		barometric_altitude_status,
		geometric_altitude_m::float8,
		geometric_altitude_status,
		velocity_mps::float8,
		heading_degrees::float8,
		vertical_rate_mps::float8,
		on_ground,
		COALESCE(origin_country, ''),
		observed_at,
		source_name
	FROM flight_states
	WHERE flight_id = $1::uuid
	  AND observed_at >= $2
	  AND observed_at <= $3
	  AND latitude IS NOT NULL
	  AND longitude IS NOT NULL
	  AND velocity_mps IS NOT NULL
	  AND heading_degrees IS NOT NULL
	  AND vertical_rate_mps IS NOT NULL
	  AND on_ground IS NOT NULL
	ORDER BY
		observed_at ASC,
		id ASC
	LIMIT $4;
`

const trajectoryPointsByAircraftSQL = `
	SELECT
		id::text,
		COALESCE(flight_id::text, ''),
		COALESCE(aircraft_id::text, ''),
		icao24,
		COALESCE(callsign, ''),
		latitude::float8,
		longitude::float8,
		barometric_altitude_m::float8,
		barometric_altitude_status,
		geometric_altitude_m::float8,
		geometric_altitude_status,
		velocity_mps::float8,
		heading_degrees::float8,
		vertical_rate_mps::float8,
		on_ground,
		COALESCE(origin_country, ''),
		observed_at,
		source_name
	FROM flight_states
	WHERE icao24 = $1
	  AND ($2 = '' OR COALESCE(callsign, '') = $2)
	  AND observed_at >= $3
	  AND observed_at <= $4
	  AND latitude IS NOT NULL
	  AND longitude IS NOT NULL
	  AND velocity_mps IS NOT NULL
	  AND heading_degrees IS NOT NULL
	  AND vertical_rate_mps IS NOT NULL
	  AND on_ground IS NOT NULL
	ORDER BY
		observed_at ASC,
		id ASC
	LIMIT $5;
`
