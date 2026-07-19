package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
)

type TransponderEvidenceAircraft struct {
	ICAO24   string `json:"icao24"`
	Callsign string `json:"callsign,omitempty"`
}

type TransponderEvidenceClassification struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type TransponderEvidenceObservation struct {
	Strength                        string `json:"strength"`
	ObservationCount                int    `json:"observation_count"`
	SpecialPurposeIndicatorObserved bool   `json:"special_purpose_indicator_observed"`
}

type TransponderEvidenceFreshness struct {
	Status                 string    `json:"status"`
	FirstObservedAt        time.Time `json:"first_observed_at"`
	LastObservedAt         time.Time `json:"last_observed_at"`
	AsOfTime               time.Time `json:"as_of_time"`
	AgeSeconds             int64     `json:"age_seconds"`
	MaximumFreshAgeSeconds int64     `json:"maximum_fresh_age_seconds"`
}

type TransponderEvidenceConfidence struct {
	Level   string   `json:"level"`
	Reasons []string `json:"reasons"`
}

type TransponderEvidenceProvenance struct {
	Fingerprint string   `json:"fingerprint"`
	SourceNames []string `json:"source_names"`
}

type TransponderEvidenceResponse struct {
	SchemaVersion string `json:"schema_version"`

	EvidenceOnly       bool `json:"evidence_only"`
	ConfirmedEmergency bool `json:"confirmed_emergency"`
	OperationalAlert   bool `json:"operational_alert"`

	Aircraft                TransponderEvidenceAircraft       `json:"aircraft"`
	ObservedTransponderCode string                            `json:"observed_transponder_code"`
	Classification          TransponderEvidenceClassification `json:"classification"`
	Observation             TransponderEvidenceObservation    `json:"observation"`
	Freshness               TransponderEvidenceFreshness      `json:"freshness"`
	Confidence              TransponderEvidenceConfidence     `json:"confidence"`
	Provenance              TransponderEvidenceProvenance     `json:"provenance"`

	MaximumClaimStrength string   `json:"maximum_claim_strength"`
	Limitations          []string `json:"limitations"`
}

func ToTransponderEvidenceResponse(
	result transponderalert.LatestEvidence,
) TransponderEvidenceResponse {
	evidence := result.Evidence

	return TransponderEvidenceResponse{
		SchemaVersion: evidence.SchemaVersion,
		EvidenceOnly:  result.EvidenceOnly,
		ConfirmedEmergency: result.
			ConfirmedEmergency,
		OperationalAlert: result.OperationalAlert,
		Aircraft: TransponderEvidenceAircraft{
			ICAO24:   evidence.ICAO24,
			Callsign: evidence.Callsign,
		},
		ObservedTransponderCode: evidence.SquawkCode,
		Classification: TransponderEvidenceClassification{
			Kind:  string(evidence.Kind),
			Label: evidence.Label,
		},
		Observation: TransponderEvidenceObservation{
			Strength: string(evidence.Strength),
			ObservationCount: evidence.
				ObservationCount,
			SpecialPurposeIndicatorObserved: evidence.
				SpecialPurposeIndicatorObserved,
		},
		Freshness: TransponderEvidenceFreshness{
			Status: string(result.FreshnessStatus),
			FirstObservedAt: evidence.
				FirstObservedAt.UTC(),
			LastObservedAt: evidence.
				LastObservedAt.UTC(),
			AsOfTime: evidence.AsOfTime.UTC(),
			AgeSeconds: int64(
				result.Age / time.Second,
			),
			MaximumFreshAgeSeconds: int64(
				result.MaximumFreshAge /
					time.Second,
			),
		},
		Confidence: TransponderEvidenceConfidence{
			Level: string(result.Confidence.Level),
			Reasons: append(
				[]string(nil),
				result.Confidence.Reasons...,
			),
		},
		Provenance: TransponderEvidenceProvenance{
			Fingerprint: evidence.Fingerprint,
			SourceNames: append(
				[]string(nil),
				evidence.SourceNames...,
			),
		},
		MaximumClaimStrength: evidence.
			MaximumClaimStrength,
		Limitations: append(
			[]string(nil),
			evidence.Limitations...,
		),
	}
}
