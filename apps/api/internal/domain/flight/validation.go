package flight

import "errors"

var (
	ErrFlightObservedRangeRequired = errors.New("flight observation range is required")
	ErrFlightObservedRangeInvalid  = errors.New("flight last-seen timestamp cannot precede first-seen timestamp")
)

func (value Flight) Validate() error {
	if value.FirstSeenAt.IsZero() || value.LastSeenAt.IsZero() {
		return ErrFlightObservedRangeRequired
	}
	if value.LastSeenAt.Before(value.FirstSeenAt) {
		return ErrFlightObservedRangeInvalid
	}
	return nil
}
