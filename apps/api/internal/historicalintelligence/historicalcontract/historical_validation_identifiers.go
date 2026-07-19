package historicalcontract

import "regexp"

var (
	airportICAOPattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
	regionCodePattern = regexp.MustCompile(
		`^[a-z0-9][a-z0-9_-]{1,31}$`,
	)
	fingerprintPattern = regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
)
