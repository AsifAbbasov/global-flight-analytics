package proximityscanner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

const timeLayout = time.RFC3339Nano

type fingerprintEnvelope struct {
	SchemaVersion    string
	PolicyVersion    string
	RegionCode       string
	SceneStatus      string
	AsOfTime         string
	SceneFingerprint string
	Candidates       []fingerprintCandidate
	Metrics          Metrics
	GraphFingerprint string
}

type fingerprintCandidate struct {
	ID                                   string
	Status                               string
	Kind                                 string
	HorizontalDistanceKilometers         float64
	VerticalSeparationMeters             *float64
	ObservationTimeDifferenceNanoseconds int64
	EffectiveHorizontalRadiusKilometers  float64
	EffectiveVerticalRadiusMeters        *float64
	VerticalFilteringApplied             bool
	ClosingRateMetersPerSecond           float64
	ConfidenceScore                      float64
}

func inputFingerprint(result Result, policy Policy) string {
	envelope := fingerprintEnvelope{
		SchemaVersion:    string(result.SchemaVersion),
		PolicyVersion:    policy.Version,
		RegionCode:       result.RegionCode,
		SceneStatus:      string(result.SceneStatus),
		AsOfTime:         result.AsOfTime.UTC().Format(timeLayout),
		SceneFingerprint: result.Provenance.SceneFingerprint,
		Candidates:       make([]fingerprintCandidate, 0, len(result.Candidates)),
		Metrics:          result.Metrics,
		GraphFingerprint: result.Graph.Provenance.InputFingerprint,
	}
	for _, candidate := range result.Candidates {
		envelope.Candidates = append(envelope.Candidates, fingerprintCandidate{
			ID:                                   candidate.ID,
			Status:                               string(candidate.Status),
			Kind:                                 string(candidate.Kind),
			HorizontalDistanceKilometers:         candidate.HorizontalDistanceKilometers,
			VerticalSeparationMeters:             cloneFloat64(candidate.VerticalSeparationMeters),
			ObservationTimeDifferenceNanoseconds: candidate.ObservationTimeDifference.Nanoseconds(),
			EffectiveHorizontalRadiusKilometers:  candidate.EffectiveHorizontalRadiusKilometers,
			EffectiveVerticalRadiusMeters:        cloneFloat64(candidate.EffectiveVerticalRadiusMeters),
			VerticalFilteringApplied:             candidate.VerticalFilteringApplied,
			ClosingRateMetersPerSecond:           candidate.ClosingRateMetersPerSecond,
			ConfidenceScore:                      candidate.Confidence.Score,
		})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
