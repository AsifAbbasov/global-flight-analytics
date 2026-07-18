package transponderalert

import "time"

const SchemaVersion = "transponder-alert-evidence-v1"

type Kind string

const (
	KindUnlawfulInterferenceCode  Kind = "unlawful_interference_code"
	KindRadioCommunicationFailure Kind = "radio_communication_failure_code"
	KindGeneralEmergencyCode      Kind = "general_emergency_code"
)

type Strength string

const (
	StrengthSingleObservation    Strength = "single_observation"
	StrengthMultipleObservations Strength = "multiple_observations"
	StrengthRepeatedObservation  Strength = "repeated_observation"
)

type Evidence struct {
	SchemaVersion string
	Fingerprint   string

	ICAO24   string
	Callsign string

	SquawkCode string
	Kind       Kind
	Label      string
	Strength   Strength

	FirstObservedAt time.Time
	LastObservedAt  time.Time
	AsOfTime        time.Time

	ObservationCount                int
	SpecialPurposeIndicatorObserved bool
	SourceNames                     []string

	MaximumClaimStrength string
	Limitations          []string
}
