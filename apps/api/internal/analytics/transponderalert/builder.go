package transponderalert

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

var (
	ErrICAO24Required = errors.New(
		"transponder alert evidence ICAO24 is required",
	)
	ErrObservationTimeRequired = errors.New(
		"transponder alert evidence observation time is required",
	)
	ErrAsOfTimeBeforeEvidence = errors.New(
		"transponder alert evidence as-of time precedes observed evidence",
	)
)

type groupKey struct {
	ICAO24 string
	Code   string
}

type accumulator struct {
	key groupKey

	callsign string
	first    time.Time
	last     time.Time
	count    int
	spi      bool
	sources  map[string]struct{}
}

func Build(
	states []flightstate.FlightState,
	asOfTime time.Time,
) ([]Evidence, error) {
	groups := make(
		map[groupKey]*accumulator,
	)

	for _, state := range states {
		code, err := flightstate.NormalizeSquawkCode(
			state.SquawkCode,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"normalize squawk for ICAO24 %q: %w",
				state.ICAO24,
				err,
			)
		}
		if !flightstate.IsSpecialTransponderCode(code) {
			continue
		}

		icao24 := strings.ToUpper(
			strings.TrimSpace(
				state.ICAO24,
			),
		)
		if icao24 == "" {
			return nil, ErrICAO24Required
		}
		if state.ObservedAt.IsZero() {
			return nil, fmt.Errorf(
				"%w: ICAO24=%s code=%s",
				ErrObservationTimeRequired,
				icao24,
				code,
			)
		}

		key := groupKey{
			ICAO24: icao24,
			Code:   code,
		}
		group := groups[key]
		if group == nil {
			group = &accumulator{
				key:     key,
				sources: make(map[string]struct{}),
			}
			groups[key] = group
		}

		observedAt := state.ObservedAt.UTC()
		if group.first.IsZero() ||
			observedAt.Before(group.first) {
			group.first = observedAt
		}
		if group.last.IsZero() ||
			observedAt.After(group.last) {
			group.last = observedAt
			if callsign := strings.TrimSpace(
				state.Callsign,
			); callsign != "" {
				group.callsign = callsign
			}
		}
		group.count++
		group.spi = group.spi ||
			state.SpecialPurposeIndicator

		source := strings.TrimSpace(
			state.SourceName,
		)
		if source == "" {
			source = "unknown"
		}
		group.sources[source] = struct{}{}
	}

	keys := make(
		[]groupKey,
		0,
		len(groups),
	)
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(
		left int,
		right int,
	) bool {
		if keys[left].ICAO24 != keys[right].ICAO24 {
			return keys[left].ICAO24 < keys[right].ICAO24
		}
		return keys[left].Code < keys[right].Code
	})

	result := make(
		[]Evidence,
		0,
		len(keys),
	)
	for _, key := range keys {
		group := groups[key]
		effectiveAsOf := asOfTime.UTC()
		if effectiveAsOf.IsZero() {
			effectiveAsOf = group.last
		}
		if effectiveAsOf.Before(group.last) {
			return nil, fmt.Errorf(
				"%w: ICAO24=%s code=%s",
				ErrAsOfTimeBeforeEvidence,
				group.key.ICAO24,
				group.key.Code,
			)
		}

		sources := make(
			[]string,
			0,
			len(group.sources),
		)
		for source := range group.sources {
			sources = append(sources, source)
		}
		sort.Strings(sources)

		kind, label := classify(
			group.key.Code,
		)
		evidence := Evidence{
			SchemaVersion: SchemaVersion,
			ICAO24:        group.key.ICAO24,
			Callsign:      group.callsign,
			SquawkCode:    group.key.Code,
			Kind:          kind,
			Label:         label,
			Strength: strength(
				group.count,
				group.last.Sub(group.first),
			),
			FirstObservedAt:                 group.first,
			LastObservedAt:                  group.last,
			AsOfTime:                        effectiveAsOf,
			ObservationCount:                group.count,
			SpecialPurposeIndicatorObserved: group.spi,
			SourceNames:                     sources,
			MaximumClaimStrength:            "observed_transponder_code_only",
			Limitations: []string{
				"An observed transponder code does not independently confirm an emergency, unlawful interference, radio failure, or incident cause.",
				"Special Purpose Indicator evidence does not upgrade the code into confirmed operational truth.",
				"Coverage is externally collected, may be incomplete, and is not first-party sensor evidence.",
				"The result is research-only and must not be used as an operational alert or directive.",
			},
		}
		evidence.Fingerprint = fingerprint(
			evidence,
		)
		result = append(result, evidence)
	}

	return result, nil
}

func classify(
	code string,
) (Kind, string) {
	switch code {
	case "7500":
		return KindUnlawfulInterferenceCode,
			"Observed transponder code associated with unlawful interference"
	case "7600":
		return KindRadioCommunicationFailure,
			"Observed transponder code associated with radio communication failure"
	default:
		return KindGeneralEmergencyCode,
			"Observed general emergency transponder code"
	}
}

func strength(
	count int,
	duration time.Duration,
) Strength {
	if count <= 1 {
		return StrengthSingleObservation
	}
	if duration >= 10*time.Second {
		return StrengthRepeatedObservation
	}
	return StrengthMultipleObservations
}

func fingerprint(
	evidence Evidence,
) string {
	builder := strings.Builder{}
	builder.WriteString(evidence.SchemaVersion)
	builder.WriteString("|")
	builder.WriteString(evidence.ICAO24)
	builder.WriteString("|")
	builder.WriteString(evidence.SquawkCode)
	builder.WriteString("|")
	builder.WriteString(
		strconv.FormatInt(
			evidence.FirstObservedAt.UnixNano(),
			10,
		),
	)
	builder.WriteString("|")
	builder.WriteString(
		strconv.FormatInt(
			evidence.LastObservedAt.UnixNano(),
			10,
		),
	)
	builder.WriteString("|")
	builder.WriteString(
		strconv.Itoa(
			evidence.ObservationCount,
		),
	)
	builder.WriteString("|")
	builder.WriteString(
		strings.Join(
			evidence.SourceNames,
			",",
		),
	)

	sum := sha256.Sum256(
		[]byte(
			builder.String(),
		),
	)
	return "sha256:" +
		hex.EncodeToString(
			sum[:],
		)
}
