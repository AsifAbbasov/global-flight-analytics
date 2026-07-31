package projectionevaluation

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var (
	ErrTruthAvailabilityEvidenceInvalid = errors.New("truth availability evidence is invalid")
	ErrTruthAvailabilityEvidenceMissing = errors.New("truth availability evidence is missing")
	ErrAmbiguousTruthTimestamp          = errors.New("conflicting truth points share the same observation timestamp")
)

type canonicalTruthPoint struct {
	point          trajectory.TrackPoint4D
	availableAt    time.Time
	evidenceSource string
}

type normalizedTruth struct {
	points                          []canonicalTruthPoint
	excludedAfterObservationCutoff  int
	excludedAfterAvailabilityCutoff int
}

type truthPoint struct {
	source    ActualPointSource
	timeValue time.Time
	latitude  float64
	longitude float64
	altitudeM *float64
}

type truthMatchStatus string

const (
	truthMatchAvailable           truthMatchStatus = "available"
	truthMatchUnavailable         truthMatchStatus = "unavailable"
	truthMatchGapExceeded         truthMatchStatus = "gap_exceeded"
	truthMatchImplausibleMovement truthMatchStatus = "implausible_movement"
)

func normalizeTruthPoints(
	item trajectory.FlightTrajectory,
	availability []TruthAvailability,
	asOfTime time.Time,
	evaluatedAt time.Time,
) (normalizedTruth, error) {
	availabilityByPointID := make(map[string]TruthAvailability, len(availability))
	for index, evidence := range availability {
		evidence.PointID = strings.TrimSpace(evidence.PointID)
		evidence.SourceName = strings.TrimSpace(evidence.SourceName)
		evidence.AvailableAt = evidence.AvailableAt.UTC()
		if evidence.PointID == "" || evidence.SourceName == "" || evidence.AvailableAt.IsZero() {
			return normalizedTruth{}, fmt.Errorf("%w at index %d", ErrTruthAvailabilityEvidenceInvalid, index)
		}
		if _, exists := availabilityByPointID[evidence.PointID]; exists {
			return normalizedTruth{}, fmt.Errorf("%w: duplicate point identifier %q", ErrTruthAvailabilityEvidenceInvalid, evidence.PointID)
		}
		availabilityByPointID[evidence.PointID] = evidence
	}

	asOfTime = asOfTime.UTC()
	evaluatedAt = evaluatedAt.UTC()
	candidates := make([]canonicalTruthPoint, 0, len(item.Points))
	result := normalizedTruth{}

	for index, point := range item.Points {
		point.ID = strings.TrimSpace(point.ID)
		observedAt := point.ObservedAt.UTC()
		if point.ID == "" || point.ObservedAt.IsZero() || observedAt.Before(asOfTime) ||
			!validLatitude(point.Latitude) || !validLongitude(point.Longitude) {
			continue
		}
		if observedAt.After(evaluatedAt) {
			result.excludedAfterObservationCutoff++
			continue
		}
		evidence, exists := availabilityByPointID[point.ID]
		if !exists {
			return normalizedTruth{}, fmt.Errorf("%w for trajectory point %q at index %d", ErrTruthAvailabilityEvidenceMissing, point.ID, index)
		}
		if evidence.AvailableAt.After(evaluatedAt) {
			result.excludedAfterAvailabilityCutoff++
			continue
		}
		point.ObservedAt = observedAt
		point.SourceName = strings.TrimSpace(point.SourceName)
		candidates = append(candidates, canonicalTruthPoint{
			point:          point,
			availableAt:    evidence.AvailableAt,
			evidenceSource: evidence.SourceName,
		})
	}

	sort.Slice(candidates, func(left, right int) bool {
		leftTime := candidates[left].point.ObservedAt
		rightTime := candidates[right].point.ObservedAt
		if !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}
		return candidates[left].point.ID < candidates[right].point.ID
	})

	result.points = make([]canonicalTruthPoint, 0, len(candidates))
	for _, candidate := range candidates {
		if len(result.points) == 0 || !candidate.point.ObservedAt.Equal(result.points[len(result.points)-1].point.ObservedAt) {
			result.points = append(result.points, candidate)
			continue
		}
		previous := result.points[len(result.points)-1]
		if !sameTruthContent(previous, candidate) {
			return normalizedTruth{}, fmt.Errorf("%w: %s", ErrAmbiguousTruthTimestamp, candidate.point.ObservedAt.Format(time.RFC3339Nano))
		}
		// Identical duplicates are harmless. Lexicographic sorting makes the
		// retained representative independent of input order.
	}

	return result, nil
}

func sameTruthContent(left, right canonicalTruthPoint) bool {
	return left.point.Latitude == right.point.Latitude &&
		left.point.Longitude == right.point.Longitude &&
		left.point.GeometricAltitudeM == right.point.GeometricAltitudeM &&
		left.point.GeometricAltitudeStatus == right.point.GeometricAltitudeStatus &&
		left.point.BarometricAltitudeM == right.point.BarometricAltitudeM &&
		left.point.BarometricAltitudeStatus == right.point.BarometricAltitudeStatus
}

func truthAt(
	points []canonicalTruthPoint,
	targetTime time.Time,
	config Config,
) (truthPoint, truthMatchStatus) {
	if len(points) == 0 || targetTime.IsZero() {
		return truthPoint{}, truthMatchUnavailable
	}
	targetTime = targetTime.UTC()
	if targetTime.Before(points[0].point.ObservedAt) || targetTime.After(points[len(points)-1].point.ObservedAt) {
		return truthPoint{}, truthMatchUnavailable
	}
	for index, item := range points {
		point := item.point
		if point.ObservedAt.Equal(targetTime) {
			return truthFromObserved(point), truthMatchAvailable
		}
		if point.ObservedAt.After(targetTime) {
			if index == 0 {
				return truthPoint{}, truthMatchUnavailable
			}
			left := points[index-1].point
			right := point
			duration := right.ObservedAt.Sub(left.ObservedAt)
			if duration > config.MaximumInterpolationGap {
				return truthPoint{}, truthMatchGapExceeded
			}
			if !plausibleTruthSegment(left, right, duration, config) {
				return truthPoint{}, truthMatchImplausibleMovement
			}
			actual, valid := interpolateTruth(left, right, targetTime)
			if !valid {
				return truthPoint{}, truthMatchUnavailable
			}
			return actual, truthMatchAvailable
		}
	}
	return truthPoint{}, truthMatchUnavailable
}

func plausibleTruthSegment(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
	duration time.Duration,
	config Config,
) bool {
	if duration <= 0 {
		return false
	}
	seconds := duration.Seconds()
	distanceM := greatCircleDistanceM(left.Latitude, left.Longitude, right.Latitude, right.Longitude)
	if !nonNegativeFinite(distanceM) || distanceM/seconds > config.MaximumTruthGroundSpeedMPS {
		return false
	}
	leftAltitudeM, leftAvailable := usableAltitude(left)
	rightAltitudeM, rightAvailable := usableAltitude(right)
	if leftAvailable && rightAvailable {
		verticalRate := abs(leftAltitudeM-rightAltitudeM) / seconds
		if !nonNegativeFinite(verticalRate) || verticalRate > config.MaximumTruthVerticalRateMPS {
			return false
		}
	}
	return true
}

func truthFromObserved(point trajectory.TrackPoint4D) truthPoint {
	result := truthPoint{
		source:    ActualPointSourceObserved,
		timeValue: point.ObservedAt.UTC(),
		latitude:  point.Latitude,
		longitude: point.Longitude,
	}
	if altitudeM, available := usableAltitude(point); available {
		result.altitudeM = float64Pointer(altitudeM)
	}
	return result
}

func interpolateTruth(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
	targetTime time.Time,
) (truthPoint, bool) {
	duration := right.ObservedAt.Sub(left.ObservedAt)
	if duration <= 0 {
		return truthPoint{}, false
	}
	fraction := float64(targetTime.Sub(left.ObservedAt)) / float64(duration)
	if !unitInterval(fraction) {
		return truthPoint{}, false
	}
	distanceM := greatCircleDistanceM(left.Latitude, left.Longitude, right.Latitude, right.Longitude)
	bearing := initialBearingDegrees(left.Latitude, left.Longitude, right.Latitude, right.Longitude)
	latitude, longitude, valid := destinationPoint(left.Latitude, left.Longitude, bearing, distanceM*fraction)
	if !valid {
		return truthPoint{}, false
	}
	result := truthPoint{
		source:    ActualPointSourceInterpolated,
		timeValue: targetTime.UTC(),
		latitude:  latitude,
		longitude: longitude,
	}
	leftAltitudeM, leftAvailable := usableAltitude(left)
	rightAltitudeM, rightAvailable := usableAltitude(right)
	if leftAvailable && rightAvailable {
		altitudeM := leftAltitudeM + (rightAltitudeM-leftAltitudeM)*fraction
		if finite(altitudeM) {
			result.altitudeM = float64Pointer(altitudeM)
		}
	}
	return result, true
}

func usableAltitude(point trajectory.TrackPoint4D) (float64, bool) {
	geometricStatus := flightstate.ResolveAltitudeStatus(point.GeometricAltitudeM, point.GeometricAltitudeStatus)
	if usableAltitudeStatus(geometricStatus) && finite(point.GeometricAltitudeM) {
		return point.GeometricAltitudeM, true
	}
	barometricStatus := flightstate.ResolveAltitudeStatus(point.BarometricAltitudeM, point.BarometricAltitudeStatus)
	if usableAltitudeStatus(barometricStatus) && finite(point.BarometricAltitudeM) {
		return point.BarometricAltitudeM, true
	}
	return 0, false
}

func usableAltitudeStatus(status flightstate.AltitudeStatus) bool {
	return status == flightstate.AltitudeStatusObserved || status == flightstate.AltitudeStatusGround
}

func float64Pointer(value float64) *float64 { return &value }
func boolPointer(value bool) *bool          { return &value }
func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
