package operationalbuilder

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type indexedSample struct {
	value      float64
	pointIndex int
}

type orderedPoint struct {
	point      trajectory.TrackPoint4D
	observedAt time.Time
	inputIndex int
}

type sampleCollection struct {
	altitudes             []float64
	velocities            []float64
	absoluteVerticalRates []float64
	headingChanges        []float64
	headingSampleCount    int

	groundObservationCount   int
	airborneObservationCount int
	groundStateCount         int
	supportingPointCount     int
	eligiblePointCount       int

	windowUnavailableCount       int
	missingTimestampCount        int
	outsideWindowCount           int
	duplicateTimestampCount      int
	invalidAltitudeCount         int
	geometricFallbackCount       int
	mixedAltitudeExcludedCount   int
	unavailableVelocityCount     int
	invalidVelocityCount         int
	unavailableVerticalRateCount int
	invalidVerticalRateCount     int
	unavailableHeadingCount      int
	invalidHeadingCount          int
	headingSequenceGapCount      int
	unavailableOnGroundCount     int
}

type altitudeAssessment struct {
	value   float64
	usable  bool
	invalid bool
}

func collectSamples(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (sampleCollection, error) {
	ordered, collection, err := collectOperationalPoints(ctx, item)
	if err != nil {
		return sampleCollection{}, err
	}
	collection.eligiblePointCount = len(ordered)
	if len(ordered) == 0 {
		return collection, nil
	}

	barometric := make([]indexedSample, 0, len(ordered))
	geometric := make([]indexedSample, 0, len(ordered))
	supporting := make([]bool, len(ordered))
	var previousHeading float64
	previousHeadingUsable := false

	for index, observed := range ordered {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return sampleCollection{}, err
			}
		}
		point := observed.point

		barometricAssessment := assessAltitude(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
			point,
		)
		geometricAssessment := assessAltitude(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
			point,
		)
		if barometricAssessment.invalid || geometricAssessment.invalid {
			collection.invalidAltitudeCount++
		}
		if barometricAssessment.usable {
			barometric = append(barometric, indexedSample{
				value:      barometricAssessment.value,
				pointIndex: index,
			})
		}
		if geometricAssessment.usable {
			geometric = append(geometric, indexedSample{
				value:      geometricAssessment.value,
				pointIndex: index,
			})
		}

		if !point.HasVelocity() {
			collection.unavailableVelocityCount++
		} else if finite(point.VelocityMPS) && point.VelocityMPS >= 0 {
			collection.velocities = append(collection.velocities, point.VelocityMPS)
			supporting[index] = true
		} else {
			collection.invalidVelocityCount++
		}

		if !point.HasVerticalRate() {
			collection.unavailableVerticalRateCount++
		} else if finite(point.VerticalRateMPS) {
			collection.absoluteVerticalRates = append(
				collection.absoluteVerticalRates,
				math.Abs(point.VerticalRateMPS),
			)
			supporting[index] = true
		} else {
			collection.invalidVerticalRateCount++
		}

		headingUsable := false
		if !point.HasHeading() {
			collection.unavailableHeadingCount++
		} else if finite(point.HeadingDegrees) &&
			point.HeadingDegrees >= 0 &&
			point.HeadingDegrees < 360 {
			headingUsable = true
			collection.headingSampleCount++
			supporting[index] = true
			if previousHeadingUsable {
				collection.headingChanges = append(
					collection.headingChanges,
					shortestHeadingChange(previousHeading, point.HeadingDegrees),
				)
			}
			previousHeading = point.HeadingDegrees
		} else {
			collection.invalidHeadingCount++
		}
		if !headingUsable {
			if previousHeadingUsable {
				collection.headingSequenceGapCount++
			}
			previousHeadingUsable = false
		} else {
			previousHeadingUsable = true
		}

		if !point.HasOnGroundState() {
			collection.unavailableOnGroundCount++
		} else {
			collection.groundStateCount++
			if point.OnGround {
				collection.groundObservationCount++
			} else {
				collection.airborneObservationCount++
			}
			supporting[index] = true
		}
	}

	selectedAltitude := barometric
	if len(selectedAltitude) == 0 {
		selectedAltitude = geometric
		collection.geometricFallbackCount = len(geometric)
	} else if len(geometric) > 0 {
		collection.mixedAltitudeExcludedCount = len(geometric)
	}
	collection.altitudes = make([]float64, 0, len(selectedAltitude))
	for _, sample := range selectedAltitude {
		collection.altitudes = append(collection.altitudes, sample.value)
		supporting[sample.pointIndex] = true
	}
	for _, usable := range supporting {
		if usable {
			collection.supportingPointCount++
		}
	}

	if err := ctx.Err(); err != nil {
		return sampleCollection{}, err
	}
	return collection, nil
}

func collectOperationalPoints(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) ([]orderedPoint, sampleCollection, error) {
	startTime, endTime, windowAvailable, err := resolveOperationalWindow(item)
	if err != nil {
		return nil, sampleCollection{}, err
	}
	collection := sampleCollection{}
	if !windowAvailable && len(item.Points) > 0 {
		collection.windowUnavailableCount = 1
	}

	ordered := make([]orderedPoint, 0, len(item.Points))
	for index, point := range item.Points {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return nil, sampleCollection{}, err
			}
		}
		observedAt := point.ObservedAt.UTC()
		if windowAvailable {
			if point.ObservedAt.IsZero() {
				collection.missingTimestampCount++
				continue
			}
			if observedAt.Before(startTime) || observedAt.After(endTime) {
				collection.outsideWindowCount++
				continue
			}
		}
		ordered = append(ordered, orderedPoint{
			point:      point,
			observedAt: observedAt,
			inputIndex: index,
		})
	}

	if windowAvailable {
		sort.SliceStable(ordered, func(left int, right int) bool {
			return orderedPointLess(ordered[left], ordered[right])
		})
		for index := 1; index < len(ordered); index++ {
			if ordered[index].observedAt.Equal(ordered[index-1].observedAt) {
				collection.duplicateTimestampCount++
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, sampleCollection{}, err
	}
	return ordered, collection, nil
}

func orderedPointLess(left orderedPoint, right orderedPoint) bool {
	if !left.observedAt.Equal(right.observedAt) {
		return left.observedAt.Before(right.observedAt)
	}
	leftPoint := left.point
	rightPoint := right.point
	leftStrings := []string{
		leftPoint.ID,
		leftPoint.FlightStateID,
		leftPoint.SourceName,
		leftPoint.FlightID,
	}
	rightStrings := []string{
		rightPoint.ID,
		rightPoint.FlightStateID,
		rightPoint.SourceName,
		rightPoint.FlightID,
	}
	for index := range leftStrings {
		if leftStrings[index] != rightStrings[index] {
			return leftStrings[index] < rightStrings[index]
		}
	}
	leftNumbers := []float64{
		leftPoint.BarometricAltitudeM,
		leftPoint.GeometricAltitudeM,
		leftPoint.VelocityMPS,
		leftPoint.HeadingDegrees,
		leftPoint.VerticalRateMPS,
		leftPoint.Latitude,
		leftPoint.Longitude,
	}
	rightNumbers := []float64{
		rightPoint.BarometricAltitudeM,
		rightPoint.GeometricAltitudeM,
		rightPoint.VelocityMPS,
		rightPoint.HeadingDegrees,
		rightPoint.VerticalRateMPS,
		rightPoint.Latitude,
		rightPoint.Longitude,
	}
	for index := range leftNumbers {
		leftBits := math.Float64bits(leftNumbers[index])
		rightBits := math.Float64bits(rightNumbers[index])
		if leftBits != rightBits {
			return leftBits < rightBits
		}
	}
	if leftPoint.OnGround != rightPoint.OnGround {
		return !leftPoint.OnGround
	}
	return left.inputIndex < right.inputIndex
}

func resolveOperationalWindow(
	item trajectory.FlightTrajectory,
) (time.Time, time.Time, bool, error) {
	if item.StartTime.IsZero() && item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, nil
	}
	if item.StartTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryStartTimeRequired
	}
	if item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryEndTimeRequired
	}
	if item.EndTime.Before(item.StartTime) {
		return time.Time{}, time.Time{}, false, ErrInvalidTrajectoryWindow
	}
	return item.StartTime.UTC(), item.EndTime.UTC(), true, nil
}

func assessAltitude(
	value float64,
	status flightstate.AltitudeStatus,
	point trajectory.TrackPoint4D,
) altitudeAssessment {
	resolvedStatus := status
	if resolvedStatus == "" {
		resolvedStatus = flightstate.ResolveAltitudeStatus(value, status)
	}
	switch resolvedStatus {
	case flightstate.AltitudeStatusObserved:
		if !finite(value) || value < 0 {
			return altitudeAssessment{invalid: true}
		}
		return altitudeAssessment{value: value, usable: true}
	case flightstate.AltitudeStatusGround:
		if value != 0 || !point.HasOnGroundState() || !point.OnGround {
			return altitudeAssessment{invalid: true}
		}
		return altitudeAssessment{value: 0, usable: true}
	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable:
		return altitudeAssessment{}
	case flightstate.AltitudeStatusInvalid:
		return altitudeAssessment{invalid: true}
	default:
		return altitudeAssessment{invalid: true}
	}
}

func shortestHeadingChange(left float64, right float64) float64 {
	delta := math.Abs(right - left)
	if delta > 180 {
		delta = 360 - delta
	}
	return delta
}

func (collection sampleCollection) limitations() []flightfeatures.FeatureLimitation {
	result := make([]flightfeatures.FeatureLimitation, 0, 24)
	appendLimitation := func(code string, message string) {
		result = append(result, flightfeatures.FeatureLimitation{Code: code, Message: message})
	}

	if collection.eligiblePointCount == 0 {
		appendLimitation(
			flightfeatures.OperationalLimitationPointEvidenceUnavailable,
			"No temporally eligible trajectory point record was available for operational feature extraction.",
		)
	}
	if collection.windowUnavailableCount > 0 {
		appendLimitation(
			flightfeatures.OperationalLimitationTemporalWindowUnavailable,
			"Trajectory start and end timestamps were unavailable, so operational evidence retained legacy input ordering without temporal-window filtering.",
		)
	}
	appendCountLimitation(&result, collection.missingTimestampCount,
		flightfeatures.OperationalLimitationPointTimestampMissing,
		"trajectory point timestamps were missing and were excluded from operational evidence")
	appendCountLimitation(&result, collection.outsideWindowCount,
		flightfeatures.OperationalLimitationPointOutsideWindow,
		"trajectory points were outside the authoritative trajectory window and were excluded from operational evidence")
	appendCountLimitation(&result, collection.duplicateTimestampCount,
		flightfeatures.OperationalLimitationDuplicatePointTimestamp,
		"duplicate point timestamps were retained as distinct observations under deterministic tie-breaking")

	if len(collection.altitudes) == 0 {
		appendLimitation(flightfeatures.OperationalLimitationAltitudeUnavailable,
			"No usable single-source altitude observation series was available.")
	}
	appendCountLimitation(&result, collection.invalidAltitudeCount,
		flightfeatures.OperationalLimitationInvalidAltitudeObservations,
		"point records contained an invalid altitude value, status, or ground-state combination and were excluded")
	if collection.geometricFallbackCount > 0 {
		appendLimitation(
			flightfeatures.OperationalLimitationGeometricAltitudeFallback,
			fmt.Sprintf("No usable barometric altitude series was available, so %d geometric altitude observations were selected under policy %q.", collection.geometricFallbackCount, flightfeatures.CurrentOperationalAltitudeSourcePolicy),
		)
	}
	if collection.mixedAltitudeExcludedCount > 0 {
		appendLimitation(
			flightfeatures.OperationalLimitationMixedAltitudeSourceExcluded,
			fmt.Sprintf("%d geometric altitude observations were excluded because a usable barometric series was selected under policy %q.", collection.mixedAltitudeExcludedCount, flightfeatures.CurrentOperationalAltitudeSourcePolicy),
		)
	}

	if len(collection.velocities) == 0 {
		appendLimitation(flightfeatures.OperationalLimitationVelocityUnavailable,
			"No available finite non-negative ground-velocity observation was usable.")
	}
	appendCountLimitation(&result, collection.unavailableVelocityCount,
		flightfeatures.OperationalLimitationVelocityMeasurementUnavailable,
		"trajectory points did not provide a velocity measurement and were not interpreted as zero")
	appendCountLimitation(&result, collection.invalidVelocityCount,
		flightfeatures.OperationalLimitationInvalidVelocityObservations,
		"available velocity observations were non-finite or negative and were excluded")

	if len(collection.absoluteVerticalRates) == 0 {
		appendLimitation(flightfeatures.OperationalLimitationVerticalRateUnavailable,
			"No available finite vertical-rate observation was usable.")
	}
	appendCountLimitation(&result, collection.unavailableVerticalRateCount,
		flightfeatures.OperationalLimitationVerticalRateMeasurementUnavailable,
		"trajectory points did not provide a vertical-rate measurement and were not interpreted as zero")
	appendCountLimitation(&result, collection.invalidVerticalRateCount,
		flightfeatures.OperationalLimitationInvalidVerticalRateObservations,
		"available vertical-rate observations were non-finite and were excluded")

	if collection.headingSampleCount == 0 {
		appendLimitation(flightfeatures.OperationalLimitationHeadingUnavailable,
			"No available finite heading in the inclusive-zero exclusive-360-degree range was usable.")
	}
	appendCountLimitation(&result, collection.unavailableHeadingCount,
		flightfeatures.OperationalLimitationHeadingMeasurementUnavailable,
		"trajectory points did not provide a heading measurement and broke heading continuity")
	appendCountLimitation(&result, collection.invalidHeadingCount,
		flightfeatures.OperationalLimitationInvalidHeadingObservations,
		"available headings were non-finite or outside the inclusive-zero exclusive-360-degree range and were excluded")
	appendCountLimitation(&result, collection.headingSequenceGapCount,
		flightfeatures.OperationalLimitationHeadingSequenceGap,
		"heading sequences were terminated by unavailable or invalid observations, preventing direct transitions across evidence gaps")

	if collection.groundStateCount == 0 {
		appendLimitation(flightfeatures.OperationalLimitationOnGroundUnavailable,
			"No explicit on-ground state was available for ground and airborne observation shares.")
	}
	appendCountLimitation(&result, collection.unavailableOnGroundCount,
		flightfeatures.OperationalLimitationOnGroundMeasurementUnavailable,
		"trajectory points did not provide an on-ground state and were excluded from share denominators")

	if collection.eligiblePointCount > 0 && collection.supportingPointCount == 0 {
		appendLimitation(
			flightfeatures.OperationalLimitationPointEvidenceUnusable,
			fmt.Sprintf("None of the %d temporally eligible trajectory points supplied a usable operational measurement.", collection.eligiblePointCount),
		)
	}
	return result
}

func appendCountLimitation(
	result *[]flightfeatures.FeatureLimitation,
	count int,
	code string,
	message string,
) {
	if count <= 0 {
		return
	}
	*result = append(*result, flightfeatures.FeatureLimitation{
		Code:    code,
		Message: fmt.Sprintf("%d %s.", count, message),
	})
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
