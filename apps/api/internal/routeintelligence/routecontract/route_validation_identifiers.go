package routecontract

import "regexp"

var (
	icao24Pattern = regexp.MustCompile(
		`^[A-F0-9]{6}$`,
	)
	icaoAirportPattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
	iataAirportPattern = regexp.MustCompile(
		`^[A-Z0-9]{3}$`,
	)
	identityKeyPattern = regexp.MustCompile(
		`^flight-identity-[0-9a-f]{64}$`,
	)
	fingerprintPattern = regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
)
