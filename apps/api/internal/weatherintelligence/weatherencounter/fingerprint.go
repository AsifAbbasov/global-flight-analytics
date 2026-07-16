package weatherencounter

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

func inputFingerprint(
	weather weathercontract.Result,
	alignment weatheralignment.Result,
	policy Policy,
	points []EncounterPoint,
) string {
	hasher := sha256.New()

	parts := []string{
		FingerprintVersion,
		weather.Provenance.InputFingerprint,
		alignment.InputFingerprint,
		policy.Version,
		formatFloat(
			policy.MinimumCompleteProfileCoverage,
		),
		formatFloat(
			policy.
				MinimumCompleteCoreMetricCoverage,
		),
	}

	for _, point := range points {
		parts = append(
			parts,
			strconv.Itoa(
				point.TrajectoryPointSequence,
			),
			strings.TrimSpace(
				point.TrajectoryPointID,
			),
			point.TrajectoryObservedAt.UTC().
				Format(time.RFC3339Nano),
			strconv.Itoa(
				point.WeatherSampleSequence,
			),
			point.WeatherValidAt.UTC().
				Format(time.RFC3339Nano),
			formatFloat(point.AlignmentScore),
			strconv.Itoa(point.FeatureCount),
		)
	}

	for _, sample := range weather.Samples {
		parts = append(
			parts,
			strconv.Itoa(sample.Sequence),
			featureValue(
				sample.Features.
					TemperatureCelsius,
			),
			featureValue(
				sample.Features.
					RelativeHumidityPercent,
			),
			featureValue(
				sample.Features.
					PrecipitationMillimeters,
			),
			featureValue(
				sample.Features.
					RainMillimeters,
			),
			featureValue(
				sample.Features.
					CloudCoverPercent,
			),
			featureValue(
				sample.Features.
					SurfacePressureHPA,
			),
			featureValue(
				sample.Features.
					WindSpeedMetersPerSecond,
			),
			featureValue(
				sample.Features.
					WindDirectionDegrees,
			),
			featureValue(
				sample.Features.
					WindGustsMetersPerSecond,
			),
		)
		if sample.Features.ConditionCode == nil {
			parts = append(
				parts,
				"condition:nil",
			)
		} else {
			parts = append(
				parts,
				sample.Features.
					ConditionCodeScheme,
				strconv.Itoa(
					*sample.Features.
						ConditionCode,
				),
			)
		}
	}

	for _, part := range parts {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}

	return "sha256:" +
		hex.EncodeToString(hasher.Sum(nil))
}

func featureValue(value *float64) string {
	if value == nil {
		return "nil"
	}
	return formatFloat(*value)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(
		value,
		'g',
		-1,
		64,
	)
}
