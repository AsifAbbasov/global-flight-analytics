package weatheralignment

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

const earthRadiusKilometers = 6371.0088

func horizontalDistanceKilometers(leftLatitude, leftLongitude, rightLatitude, rightLongitude float64) float64 {
	leftLatRadians := degreesToRadians(leftLatitude)
	rightLatRadians := degreesToRadians(rightLatitude)
	latitudeDelta := degreesToRadians(rightLatitude - leftLatitude)
	longitudeDelta := degreesToRadians(rightLongitude - leftLongitude)

	sineLatitude := math.Sin(latitudeDelta / 2)
	sineLongitude := math.Sin(longitudeDelta / 2)
	haversine := sineLatitude*sineLatitude +
		math.Cos(leftLatRadians)*math.Cos(rightLatRadians)*sineLongitude*sineLongitude
	centralAngle := 2 * math.Atan2(math.Sqrt(haversine), math.Sqrt(1-haversine))
	return earthRadiusKilometers * centralAngle
}

func absoluteDuration(left, right time.Time) time.Duration {
	difference := left.Sub(right)
	if difference < 0 {
		return -difference
	}
	return difference
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func inputFingerprint(
	flightTrajectory trajectory.FlightTrajectory,
	weather weathercontract.Result,
	trust weathertrust.Result,
	policy Policy,
) string {
	hasher := sha256.New()
	parts := []string{
		FingerprintVersion,
		strings.TrimSpace(flightTrajectory.ID),
		weather.Provenance.InputFingerprint,
		trust.InputFingerprint,
		policy.Version,
		formatFloat(policy.MaximumHorizontalDistanceKilometers),
		policy.MaximumTemporalDistance.String(),
		formatFloat(policy.MaximumVerticalDistanceMeters),
		formatFloat(policy.MinimumMatchScore),
		formatFloat(policy.Weights.Horizontal),
		formatFloat(policy.Weights.Temporal),
		formatFloat(policy.Weights.Vertical),
	}

	for sequence, point := range flightTrajectory.Points {
		parts = append(parts,
			strconv.Itoa(sequence),
			strings.TrimSpace(point.ID),
			formatFloat(point.Latitude),
			formatFloat(point.Longitude),
			formatFloat(point.GeometricAltitudeM),
			string(point.GeometricAltitudeStatus),
			formatFloat(point.BarometricAltitudeM),
			string(point.BarometricAltitudeStatus),
			strconv.FormatBool(point.OnGround),
			point.ObservedAt.UTC().Format(time.RFC3339Nano),
		)
	}

	for _, sample := range weather.Samples {
		parts = append(parts,
			strconv.Itoa(sample.Sequence),
			formatFloat(sample.Position.Latitude),
			formatFloat(sample.Position.Longitude),
			string(sample.Position.VerticalReference),
			sample.ValidAt.UTC().Format(time.RFC3339Nano),
			sample.AvailableAt.UTC().Format(time.RFC3339Nano),
		)
		if sample.Position.AltitudeMeters != nil {
			parts = append(parts, formatFloat(*sample.Position.AltitudeMeters))
		} else {
			parts = append(parts, "altitude:nil")
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
