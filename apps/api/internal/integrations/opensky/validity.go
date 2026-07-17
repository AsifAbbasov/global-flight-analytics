package opensky

import (
	"errors"
	"fmt"
	"time"
)

const MaximumProviderFieldAge = 15 * time.Second

type PositionValidity string

const (
	PositionValidityProviderValid PositionValidity = "provider_valid"
	PositionValidityUnavailable   PositionValidity = "unavailable"
	PositionValidityStale         PositionValidity = "stale"
	PositionValidityInvalid       PositionValidity = "invalid"
)

var (
	ErrSnapshotTimeRequired      = errors.New("OpenSky snapshot time is required")
	ErrLastContactAfterSnapshot  = errors.New("OpenSky last contact is after snapshot time")
	ErrPositionTimeAfterSnapshot = errors.New("OpenSky position time is after snapshot time")
)

type StateVectorValidity struct {
	SnapshotTime            time.Time        `json:"snapshot_time"`
	LastContact             time.Time        `json:"last_contact"`
	TimePosition            *time.Time       `json:"time_position"`
	LastContactAgeSeconds   float64          `json:"last_contact_age_seconds"`
	PositionAgeSeconds      *float64         `json:"position_age_seconds"`
	PositionValidity        PositionValidity `json:"position_validity"`
	PositionUsable          bool             `json:"position_usable"`
	LastContactWithinWindow bool             `json:"last_contact_within_window"`
	Limitations             []string         `json:"limitations"`
}

func EvaluateStateVectorValidity(
	state StateVector,
) (StateVectorValidity, error) {
	if state.SnapshotTime.IsZero() {
		return StateVectorValidity{}, ErrSnapshotTimeRequired
	}

	lastContactAge := state.SnapshotTime.Sub(state.LastContact)
	if lastContactAge < 0 {
		return StateVectorValidity{}, fmt.Errorf(
			"%w: snapshot=%s last_contact=%s",
			ErrLastContactAfterSnapshot,
			state.SnapshotTime.Format(time.RFC3339),
			state.LastContact.Format(time.RFC3339),
		)
	}

	result := StateVectorValidity{
		SnapshotTime:            state.SnapshotTime.UTC(),
		LastContact:             state.LastContact.UTC(),
		TimePosition:            cloneTime(state.TimePosition),
		LastContactAgeSeconds:   lastContactAge.Seconds(),
		LastContactWithinWindow: lastContactAge <= MaximumProviderFieldAge,
		PositionValidity:        PositionValidityUnavailable,
		Limitations: []string{
			"OpenSky State Vector fields may have different source timestamps and must not be treated as one simultaneous sensor packet.",
		},
	}

	if !result.LastContactWithinWindow {
		result.Limitations = append(
			result.Limitations,
			"The last transponder contact exceeds the provider fifteen-second validity window.",
		)
	}

	if state.TimePosition == nil ||
		state.Latitude == nil ||
		state.Longitude == nil {
		result.Limitations = append(
			result.Limitations,
			"Position is unavailable and must not be reconstructed as an observed point.",
		)
		return result, nil
	}

	positionAge := state.SnapshotTime.Sub(*state.TimePosition)
	if positionAge < 0 {
		return StateVectorValidity{}, fmt.Errorf(
			"%w: snapshot=%s time_position=%s",
			ErrPositionTimeAfterSnapshot,
			state.SnapshotTime.Format(time.RFC3339),
			state.TimePosition.Format(time.RFC3339),
		)
	}

	positionAgeSeconds := positionAge.Seconds()
	result.PositionAgeSeconds = &positionAgeSeconds
	if positionAge > MaximumProviderFieldAge {
		result.PositionValidity = PositionValidityStale
		result.Limitations = append(
			result.Limitations,
			"Position exceeds the provider fifteen-second reuse window and is blocked from observed-position analytics.",
		)
		return result, nil
	}

	result.PositionValidity = PositionValidityProviderValid
	result.PositionUsable = true
	result.Limitations = append(
		result.Limitations,
		"Position is within the provider validity window but may be a reused last-known position rather than a newly received position message.",
	)
	return result, nil
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}
