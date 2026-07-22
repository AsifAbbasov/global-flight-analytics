package aircraft

import (
	"errors"
	"strings"
)

var ErrAircraftICAO24Invalid = errors.New("aircraft ICAO24 must contain six hexadecimal characters")

func (value Aircraft) Validate() error {
	normalized := strings.ToLower(strings.TrimSpace(value.ICAO24))
	if len(normalized) != 6 {
		return ErrAircraftICAO24Invalid
	}
	for _, character := range normalized {
		if (character < '0' || character > '9') &&
			(character < 'a' || character > 'f') {
			return ErrAircraftICAO24Invalid
		}
	}
	return nil
}
