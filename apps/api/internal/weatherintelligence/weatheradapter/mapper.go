// Package weatheradapter maps existing provider-domain weather payloads
// into the canonical Weather Feature Contract.
package weatheradapter

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

const (
	Version = "weather-open-meteo-current-snapshot-adapter-v1"

	CurrentSnapshotDataset = "open_meteo_current_weather"

	WMOConditionCodeScheme = "wmo_weather_interpretation_code"

	FingerprintVersion = "weather-open-meteo-current-snapshot-fingerprint-v1"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"trajectory identifier is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"weather adapter as-of time is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"weather adapter generated-at time is invalid",
	)
	ErrProviderMismatch = errors.New(
		"weather snapshot is not from Open-Meteo",
	)
	ErrSnapshotTimeInvalid = errors.New(
		"weather snapshot times are invalid",
	)
	ErrFutureSnapshotEvidence = errors.New(
		"weather snapshot was not available at the as-of time",
	)
	ErrMappedResultInvalid = errors.New(
		"mapped weather feature result is invalid",
	)
)

type Request struct {
	TrajectoryID string
	AsOfTime     time.Time
	GeneratedAt  time.Time

	Snapshot domainweather.CurrentSnapshot
}

func MapOpenMeteoCurrentSnapshot(
	request Request,
) (weathercontract.Result, error) {
	trajectoryID := strings.TrimSpace(
		request.TrajectoryID,
	)
	if trajectoryID == "" {
		return weathercontract.Result{},
			ErrTrajectoryIDRequired
	}
	if request.AsOfTime.IsZero() {
		return weathercontract.Result{},
			ErrAsOfTimeRequired
	}

	asOfTime := request.AsOfTime.UTC()
	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(asOfTime) {
		return weathercontract.Result{},
			ErrGeneratedAtInvalid
	}

	snapshot := request.Snapshot
	if strings.TrimSpace(snapshot.Provider) !=
		domainweather.ProviderOpenMeteo {
		return weathercontract.Result{},
			ErrProviderMismatch
	}
	if snapshot.ObservedAt.IsZero() ||
		snapshot.RetrievedAt.IsZero() ||
		snapshot.RetrievedAt.Before(
			snapshot.ObservedAt,
		) ||
		generatedAt.Before(snapshot.RetrievedAt) {
		return weathercontract.Result{},
			ErrSnapshotTimeInvalid
	}
	if snapshot.RetrievedAt.After(asOfTime) {
		return weathercontract.Result{},
			ErrFutureSnapshotEvidence
	}

	result := weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusLimited,
		TrajectoryID:  trajectoryID,
		AsOfTime:      asOfTime,
		Samples: []weathercontract.Sample{
			mapSample(snapshot),
		},
		Confidence: weathercontract.Confidence{
			Score: 0.55,
			Level: weathercontract.
				ConfidenceLevelMedium,
			Reasons: []weathercontract.ConfidenceReason{
				{
					Code:         "provider_snapshot_available",
					Message:      "An Open-Meteo current weather snapshot is available.",
					Contribution: 0.70,
				},
				{
					Code:         "surface_only_weather",
					Message:      "The current snapshot represents surface weather rather than flight-level weather.",
					Contribution: -0.10,
				},
				{
					Code:         "availability_time_approximated",
					Message:      "Provider publication time is unavailable, so retrieval time is used as the availability boundary.",
					Contribution: -0.05,
				},
			},
		},
		Limitations: []weathercontract.Limitation{
			{
				Code:    "surface_weather_not_flight_level",
				Message: "The mapped snapshot is surface weather and must not be interpreted as weather at aircraft altitude.",
				Scope:   "vertical_alignment",
			},
			{
				Code:    "provider_availability_time_unavailable",
				Message: "The provider publication time is not exposed by the existing snapshot model; retrieval time is used conservatively.",
				Scope:   "temporal_provenance",
			},
			{
				Code:    "trajectory_alignment_not_applied",
				Message: "The snapshot has not yet been aligned to trajectory points in space, altitude, and time.",
				Scope:   "trajectory_alignment",
			},
		},
		Explanations: []weathercontract.Explanation{
			{
				Code:    "open_meteo_snapshot_mapped",
				Message: "The existing Open-Meteo current snapshot was mapped into the canonical Weather Feature Contract.",
			},
			{
				Code:    "weather_context_only",
				Message: "Weather is contextual evidence and is not proof of pilot intent, controller intent, rerouting reason, or maneuver cause.",
			},
		},
		ScopeGuard: weathercontract.
			ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint: snapshotFingerprint(
				trajectoryID,
				asOfTime,
				snapshot,
			),
			SourceNames: []string{
				domainweather.ProviderOpenMeteo,
			},
			LatestAvailableAt: snapshot.
				RetrievedAt.UTC(),
		},
		GeneratedAt: generatedAt,
	}

	report := weathercontract.Validate(result)
	if report.Status !=
		weathercontract.ValidationStatusValid {
		return weathercontract.Result{},
			fmt.Errorf(
				"%w: issues=%v",
				ErrMappedResultInvalid,
				report.Issues,
			)
	}

	return result.Clone(), nil
}

func mapSample(
	snapshot domainweather.CurrentSnapshot,
) weathercontract.Sample {
	temperature := snapshot.TemperatureCelsius
	humidity := float64(
		snapshot.RelativeHumidityPercent,
	)
	precipitation :=
		snapshot.PrecipitationMillimeters
	rain := snapshot.RainMillimeters
	conditionCode := snapshot.WeatherCode
	cloudCover := float64(
		snapshot.CloudCoverPercent,
	)
	pressure := snapshot.SurfacePressureHPA
	windSpeed :=
		snapshot.WindSpeedMetersPerSecond
	windDirection := float64(
		snapshot.WindDirectionDegrees,
	)
	windGusts :=
		snapshot.WindGustsMetersPerSecond

	return weathercontract.Sample{
		Sequence: 0,
		Position: weathercontract.Position{
			Latitude:  snapshot.Latitude,
			Longitude: snapshot.Longitude,
			VerticalReference: weathercontract.
				VerticalReferenceSurface,
		},
		Source: weathercontract.Source{
			Provider: domainweather.
				ProviderOpenMeteo,
			Dataset: CurrentSnapshotDataset,
			EvidenceKind: weathercontract.
				EvidenceKindAnalysis,
		},
		Features: weathercontract.FeatureVector{
			TemperatureCelsius:       &temperature,
			RelativeHumidityPercent:  &humidity,
			PrecipitationMillimeters: &precipitation,
			RainMillimeters:          &rain,
			CloudCoverPercent:        &cloudCover,
			SurfacePressureHPA:       &pressure,
			WindSpeedMetersPerSecond: &windSpeed,
			WindDirectionDegrees:     &windDirection,
			WindGustsMetersPerSecond: &windGusts,
			ConditionCode:            &conditionCode,
			ConditionCodeScheme:      WMOConditionCodeScheme,
		},
		ValidAt: snapshot.ObservedAt.UTC(),
		AvailableAt: snapshot.
			RetrievedAt.UTC(),
		RetrievedAt: snapshot.
			RetrievedAt.UTC(),
	}
}

func snapshotFingerprint(
	trajectoryID string,
	asOfTime time.Time,
	snapshot domainweather.CurrentSnapshot,
) string {
	hasher := sha256.New()

	parts := []string{
		FingerprintVersion,
		strings.TrimSpace(trajectoryID),
		asOfTime.UTC().Format(
			time.RFC3339Nano,
		),
		strings.TrimSpace(snapshot.Provider),
		formatFloat(snapshot.Latitude),
		formatFloat(snapshot.Longitude),
		snapshot.ObservedAt.UTC().Format(
			time.RFC3339Nano,
		),
		snapshot.RetrievedAt.UTC().Format(
			time.RFC3339Nano,
		),
		formatFloat(
			snapshot.TemperatureCelsius,
		),
		strconv.Itoa(
			snapshot.RelativeHumidityPercent,
		),
		formatFloat(
			snapshot.PrecipitationMillimeters,
		),
		formatFloat(
			snapshot.RainMillimeters,
		),
		strconv.Itoa(snapshot.WeatherCode),
		strconv.Itoa(
			snapshot.CloudCoverPercent,
		),
		formatFloat(
			snapshot.SurfacePressureHPA,
		),
		formatFloat(
			snapshot.WindSpeedMetersPerSecond,
		),
		strconv.Itoa(
			snapshot.WindDirectionDegrees,
		),
		formatFloat(
			snapshot.WindGustsMetersPerSecond,
		),
	}

	for _, part := range parts {
		_, _ = hasher.Write(
			[]byte(part),
		)
		_, _ = hasher.Write(
			[]byte{0},
		)
	}

	return "sha256:" +
		hex.EncodeToString(hasher.Sum(nil))
}

func formatFloat(value float64) string {
	if math.IsNaN(value) {
		return "nan"
	}
	if math.IsInf(value, 1) {
		return "+inf"
	}
	if math.IsInf(value, -1) {
		return "-inf"
	}

	return strconv.FormatFloat(
		value,
		'g',
		-1,
		64,
	)
}
