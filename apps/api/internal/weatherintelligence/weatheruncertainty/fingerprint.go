package weatheruncertainty

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

func inputFingerprint(
	projection projectioncontract.Result,
	trust weathertrust.Result,
	encounter weatherencounter.Result,
	policy Policy,
) string {
	hasher := sha256.New()

	parts := []string{
		FingerprintVersion,
		projection.Provenance.InputFingerprint,
		trust.InputFingerprint,
		encounter.InputFingerprint,
		policy.Version,
		formatFloat(policy.MaximumUncertaintyMultiplier),
		formatFloat(policy.MaximumConfidenceReduction),
		formatFloat(policy.NearTermEffectFraction),
		formatFloat(policy.WindSpeedReferenceMetersPerSecond),
		formatFloat(policy.WindSpeedHighMetersPerSecond),
		formatFloat(policy.WindGustReferenceMetersPerSecond),
		formatFloat(policy.WindGustHighMetersPerSecond),
		formatFloat(policy.PrecipitationReferenceMillimeters),
		formatFloat(policy.PrecipitationHighMillimeters),
		formatFloat(policy.CloudCoverReferencePercent),
		formatFloat(policy.CloudCoverHighPercent),
		formatFloat(policy.Weights.WindSpeed),
		formatFloat(policy.Weights.WindGust),
		formatFloat(policy.Weights.Precipitation),
		formatFloat(policy.Weights.CloudCover),
		formatFloat(policy.Weights.EvidenceQuality),
	}

	for _, point := range projection.Points {
		parts = append(
			parts,
			strconv.Itoa(point.Sequence),
			point.ForecastTime.UTC().Format(time.RFC3339Nano),
			formatFloat(point.Uncertainty.HorizontalRadiusM),
			formatFloat(point.Confidence.Score),
		)
		if point.Uncertainty.VerticalRadiusM == nil {
			parts = append(parts, "vertical:nil")
		} else {
			parts = append(parts, formatFloat(*point.Uncertainty.VerticalRadiusM))
		}
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
