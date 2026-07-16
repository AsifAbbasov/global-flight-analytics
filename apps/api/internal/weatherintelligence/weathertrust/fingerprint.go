package weathertrust

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

func inputFingerprint(input weathercontract.Result, policy Policy) string {
	hasher := sha256.New()
	parts := []string{
		FingerprintVersion,
		input.Provenance.InputFingerprint,
		string(input.SchemaVersion),
		string(input.Status),
		strings.TrimSpace(input.TrajectoryID),
		input.AsOfTime.UTC().Format(time.RFC3339Nano),
		formatFloat(input.Confidence.Score),
		string(input.Confidence.Level),
		policy.Version,
		policy.MaximumObservationAge.String(),
		policy.MaximumAnalysisAge.String(),
		policy.MaximumForecastLead.String(),
		strconv.Itoa(policy.MinimumFeatureCount),
		strconv.Itoa(policy.TargetFeatureCount),
		formatFloat(policy.MinimumUsableConfidence),
		formatFloat(policy.MinimumAllowedConfidence),
		formatFloat(policy.MinimumUsableScore),
		formatFloat(policy.MinimumAllowedScore),
		formatFloat(policy.Weights.ContractConfidence),
		formatFloat(policy.Weights.TemporalFreshness),
		formatFloat(policy.Weights.FeatureCompleteness),
		formatFloat(policy.Weights.VerticalApplicability),
	}
	for _, sample := range input.Samples {
		parts = append(parts,
			strconv.Itoa(sample.Sequence),
			sample.Source.Provider,
			sample.Source.Dataset,
			string(sample.Source.EvidenceKind),
			string(sample.Position.VerticalReference),
			sample.ValidAt.UTC().Format(time.RFC3339Nano),
			sample.AvailableAt.UTC().Format(time.RFC3339Nano),
			sample.RetrievedAt.UTC().Format(time.RFC3339Nano),
			strconv.Itoa(sample.Features.PresentCount()),
		)
	}
	for _, part := range parts {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil))
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
