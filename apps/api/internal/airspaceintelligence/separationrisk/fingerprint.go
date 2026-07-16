package separationrisk

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

const timeLayout = time.RFC3339Nano

type fingerprintEnvelope struct {
	SchemaVersion   string
	PolicyVersion   string
	RegionCode      string
	AsOfTime        string
	ScanFingerprint string
	Assessments     []fingerprintAssessment
	Metrics         Metrics
}

type fingerprintAssessment struct {
	CandidateID                          string
	Status                               string
	Level                                string
	Kind                                 string
	HorizontalDistanceKilometers         float64
	VerticalSeparationMeters             *float64
	ObservationTimeDifferenceNanoseconds int64
	ClosingRateMetersPerSecond           float64
	HorizontalRadiusRatio                *float64
	VerticalRadiusRatio                  *float64
	RiskScore                            *float64
	ConfidenceScore                      float64
}

func inputFingerprint(result Result, policy Policy) string {
	envelope := fingerprintEnvelope{
		SchemaVersion:   string(result.SchemaVersion),
		PolicyVersion:   policy.Version,
		RegionCode:      result.RegionCode,
		AsOfTime:        result.AsOfTime.UTC().Format(timeLayout),
		ScanFingerprint: result.Provenance.ScanFingerprint,
		Assessments:     make([]fingerprintAssessment, 0, len(result.Assessments)),
		Metrics:         result.Metrics,
	}
	for _, assessment := range result.Assessments {
		envelope.Assessments = append(envelope.Assessments, fingerprintAssessment{
			CandidateID:                          assessment.CandidateID,
			Status:                               string(assessment.Status),
			Level:                                string(assessment.Level),
			Kind:                                 string(assessment.Kind),
			HorizontalDistanceKilometers:         assessment.HorizontalDistanceKilometers,
			VerticalSeparationMeters:             cloneFloat64(assessment.VerticalSeparationMeters),
			ObservationTimeDifferenceNanoseconds: assessment.ObservationTimeDifference.Nanoseconds(),
			ClosingRateMetersPerSecond:           assessment.ClosingRateMetersPerSecond,
			HorizontalRadiusRatio:                cloneFloat64(assessment.HorizontalRadiusRatio),
			VerticalRadiusRatio:                  cloneFloat64(assessment.VerticalRadiusRatio),
			RiskScore:                            cloneFloat64(assessment.RiskScore),
			ConfidenceScore:                      assessment.Confidence.Score,
		})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
