package postgres

const airportSelectColumns = `
	a.id::text,
	COALESCE(a.icao_code, ''),
	COALESCE(a.iata_code, ''),
	a.name,
	COALESCE(a.city, ''),
	COALESCE(c.name, ''),
	a.latitude,
	a.longitude,
	a.elevation_ft,
	COALESCE(a.timezone, ''),
	COALESCE(ap.description, '')
`

const airportReadJoins = `
	FROM airports AS a
	LEFT JOIN countries AS c ON c.id = a.country_id
	LEFT JOIN airport_profiles AS ap ON ap.airport_id = a.id
`

const airportListFirstPageQuery = `
	SELECT ` + airportSelectColumns + airportReadJoins + `
	ORDER BY a.name ASC, a.id ASC
	LIMIT $1;
`

const airportListAfterCursorQuery = `
	SELECT ` + airportSelectColumns + airportReadJoins + `
	WHERE a.name > $1
		OR (
			a.name = $1
			AND a.id > $2::uuid
		)
	ORDER BY a.name ASC, a.id ASC
	LIMIT $3;
`

const airportByICAOQuery = `
	SELECT ` + airportSelectColumns + airportReadJoins + `
	WHERE a.icao_code = $1
	ORDER BY a.id ASC
	LIMIT 1;
`
