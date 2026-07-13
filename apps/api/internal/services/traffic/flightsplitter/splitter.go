package flightsplitter

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const identityKeyPrefix = "flight-identity-"

type groupState struct {
	observations          []Observation
	startReason           trajectory.FlightSplitReason
	lastSourceFlightID    string
	lastCallsign          string
	seenAirborne          bool
	seenGroundAfterFlight bool
}

func Split(observations []Observation) []Group {
	grouped := groupByAircraft(observations)
	icao24Values := make([]string, 0, len(grouped))
	for icao24 := range grouped {
		icao24Values = append(icao24Values, icao24)
	}
	sort.Strings(icao24Values)

	result := make([]Group, 0)
	for _, icao24 := range icao24Values {
		aircraftObservations := copyAndSort(grouped[icao24])
		result = append(result, splitAircraft(icao24, aircraftObservations)...)
	}

	return result
}

func splitAircraft(icao24 string, observations []Observation) []Group {
	if len(observations) == 0 {
		return nil
	}

	state := newGroupState(
		trajectory.FlightSplitReasonInitialObservation,
		observations[0],
	)
	groups := make([]Group, 0, 1)

	for _, observation := range observations[1:] {
		reason := splitReason(state, observation)
		if reason != "" {
			groups = append(groups, finalizeGroup(icao24, state))
			state = newGroupState(reason, observation)
			continue
		}

		state.observations = append(state.observations, observation)
		state.observe(observation)
	}

	groups = append(groups, finalizeGroup(icao24, state))
	return groups
}

func newGroupState(
	reason trajectory.FlightSplitReason,
	observation Observation,
) *groupState {
	state := &groupState{
		observations: []Observation{observation},
		startReason:  reason,
	}
	state.observe(observation)
	return state
}

func (state *groupState) observe(observation Observation) {
	sourceFlightID := normalizeSourceFlightID(observation.State.FlightID)
	if sourceFlightID != "" {
		state.lastSourceFlightID = sourceFlightID
	}

	callsign := normalizeCallsign(observation.State.Callsign)
	if callsign != "" {
		state.lastCallsign = callsign
	}

	if observation.State.OnGround {
		if state.seenAirborne {
			state.seenGroundAfterFlight = true
		}
		return
	}

	state.seenAirborne = true
}

func splitReason(
	state *groupState,
	observation Observation,
) trajectory.FlightSplitReason {
	currentSourceFlightID := normalizeSourceFlightID(observation.State.FlightID)

	if state.lastSourceFlightID != "" &&
		currentSourceFlightID != "" &&
		currentSourceFlightID != state.lastSourceFlightID {
		return trajectory.FlightSplitReasonSourceFlightIDChanged
	}

	// A source-provided flight identifier is stronger evidence than a callsign
	// or an observed ground cycle. We only use weaker evidence while no explicit
	// source flight identity is active.
	if state.lastSourceFlightID == "" && currentSourceFlightID == "" {
		currentCallsign := normalizeCallsign(observation.State.Callsign)
		if state.lastCallsign != "" &&
			currentCallsign != "" &&
			currentCallsign != state.lastCallsign {
			return trajectory.FlightSplitReasonCallsignChanged
		}

		if state.seenAirborne &&
			state.seenGroundAfterFlight &&
			!observation.State.OnGround {
			return trajectory.FlightSplitReasonGroundCycle
		}
	}

	return ""
}

func finalizeGroup(icao24 string, state *groupState) Group {
	identityKey, identityBasis := buildIdentity(
		icao24,
		state.observations,
	)

	observations := append([]Observation(nil), state.observations...)

	return Group{
		ICAO24:        icao24,
		IdentityKey:   identityKey,
		IdentityBasis: identityBasis,
		SplitReason:   state.startReason,
		Observations:  observations,
	}
}

func buildIdentity(
	icao24 string,
	observations []Observation,
) (string, trajectory.FlightIdentityBasis) {
	for _, observation := range observations {
		sourceFlightID := normalizeSourceFlightID(observation.State.FlightID)
		if sourceFlightID != "" {
			return hashedIdentityKey(
				string(trajectory.FlightIdentityBasisSourceFlightID),
				icao24,
				sourceFlightID,
			), trajectory.FlightIdentityBasisSourceFlightID
		}
	}

	startTime := observations[0].State.ObservedAt.UTC()
	callsign := ""
	for _, observation := range observations {
		callsign = normalizeCallsign(observation.State.Callsign)
		if callsign != "" {
			break
		}
	}

	if callsign != "" {
		return hashedIdentityKey(
			string(trajectory.FlightIdentityBasisCallsignAndStartTime),
			icao24,
			callsign,
			formatIdentityTime(startTime),
		), trajectory.FlightIdentityBasisCallsignAndStartTime
	}

	return hashedIdentityKey(
		string(trajectory.FlightIdentityBasisAircraftAndStartTime),
		icao24,
		formatIdentityTime(startTime),
	), trajectory.FlightIdentityBasisAircraftAndStartTime
}

func hashedIdentityKey(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return identityKeyPrefix + hex.EncodeToString(sum[:])
}

func formatIdentityTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func groupByAircraft(observations []Observation) map[string][]Observation {
	result := make(map[string][]Observation)
	for _, observation := range observations {
		icao24 := normalizeICAO24(observation.State.ICAO24)
		if icao24 == "" {
			continue
		}

		observation.State.ICAO24 = icao24
		result[icao24] = append(result[icao24], observation)
	}
	return result
}

func copyAndSort(observations []Observation) []Observation {
	result := append([]Observation(nil), observations...)
	sort.SliceStable(result, func(left int, right int) bool {
		leftState := result[left].State
		rightState := result[right].State

		if !leftState.ObservedAt.Equal(rightState.ObservedAt) {
			return leftState.ObservedAt.Before(rightState.ObservedAt)
		}
		if leftState.ID != rightState.ID {
			return leftState.ID < rightState.ID
		}
		if leftState.FlightID != rightState.FlightID {
			return leftState.FlightID < rightState.FlightID
		}
		return leftState.Callsign < rightState.Callsign
	})
	return result
}

func normalizeICAO24(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeCallsign(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeSourceFlightID(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if !isUUID(normalized) {
		return ""
	}

	return normalized
}

func isUUID(value string) bool {
	if len(value) != 36 {
		return false
	}

	for index, character := range value {
		switch index {
		case 8, 13, 18, 23:
			if character != '-' {
				return false
			}

		default:
			if !isHexadecimal(character) {
				return false
			}
		}
	}

	return true
}

func isHexadecimal(character rune) bool {
	return character >= '0' && character <= '9' ||
		character >= 'a' && character <= 'f'
}
