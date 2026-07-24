package trajectoryeligibility

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestEvaluateAllowsHealthyTrajectoryForEveryCapability(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		decision := requireDecision(
			t,
			evaluation,
			capability,
		)

		if !decision.Allowed {
			t.Fatalf(
				"expected %s to be allowed, got reasons %v",
				capability,
				decision.Reasons,
			)
		}

		if !evaluation.Permissions.Allowed(
			capability,
		) {
			t.Fatalf(
				"expected permission flag for %s",
				capability,
			)
		}
	}
}

func TestEvaluateAllowsTrafficMetricsButRejectsWeakAircraftOnlyIdentity(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.IdentityBasis =
		trajectory.FlightIdentityBasisAircraftAndStartTime
	item.Callsign = ""

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireAllowed(
		t,
		evaluation,
		CapabilityTrafficMetrics,
	)

	for _, capability := range []Capability{
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonIdentityNotReliable,
		)
	}
}

func TestEvaluateRejectsMissingIdentityOnlyWhereRequired(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.IdentityKey = ""
	item.IdentityBasis = ""
	item.SplitReason = ""

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireAllowed(
		t,
		evaluation,
		CapabilityTrafficMetrics,
	)

	for _, capability := range []Capability{
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonMissingIdentity,
		)
	}
}

func TestEvaluateAppliesCapabilitySpecificQualityThresholds(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.QualityScore = 0.55

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityHistoricalAggregation,
	} {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}

	for _, capability := range []Capability{
		CapabilityRouteInference,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonLowQualityScore,
		)
	}
}

func TestEvaluateRejectsNonFiniteQualityScore(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.QualityScore = math.NaN()

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonLowQualityScore,
		)
	}
}

func TestEvaluateAppliesCapabilitySpecificCoverageGapLimits(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.CoverageGapCount = 2
	item.CoverageGaps = []trajectory.CoverageGap{
		{},
		{},
	}

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityHistoricalAggregation,
	} {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}

	for _, capability := range []Capability{
		CapabilityRouteInference,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonTooManyCoverageGaps,
		)
	}
}

func TestEvaluateAppliesCapabilitySpecificFreshnessLimits(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.StartTime =
		now.Add(-15 * time.Minute)
	item.EndTime =
		now.Add(-10 * time.Minute)
	item.DurationSeconds = 300
	item.Points = eligibilityPoints(
		item.EndTime,
	)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireDeniedWithReason(
		t,
		evaluation,
		CapabilityTrafficMetrics,
		ReasonStaleObservations,
	)
	requireAllowed(
		t,
		evaluation,
		CapabilityAirportActivity,
	)
	requireAllowed(
		t,
		evaluation,
		CapabilityRouteInference,
	)
	requireAllowed(
		t,
		evaluation,
		CapabilityHistoricalAggregation,
	)
	requireDeniedWithReason(
		t,
		evaluation,
		CapabilityProjection,
		ReasonStaleObservations,
	)
}

func TestEvaluateRejectsMissingEvaluationTimeForLiveCapabilities(
	t *testing.T,
) {
	item := healthyTrajectory(
		eligibilityTestTime(),
	)

	evaluation := NewDefault().Evaluate(
		item,
		time.Time{},
	)

	for _, capability := range []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonEvaluationTimeMissing,
		)
	}

	requireAllowed(
		t,
		evaluation,
		CapabilityRouteInference,
	)
	requireAllowed(
		t,
		evaluation,
		CapabilityHistoricalAggregation,
	)
}

func TestEvaluateRejectsProjectionWithoutRecentAltitude(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)

	for index := range item.Points {
		item.Points[index].BarometricAltitudeM = 0
		item.Points[index].BarometricAltitudeStatus =
			flightstate.AltitudeStatusUnavailable
		item.Points[index].GeometricAltitudeM = 0
		item.Points[index].GeometricAltitudeStatus =
			flightstate.AltitudeStatusUnavailable
	}

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
	} {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}

	requireDeniedWithReason(
		t,
		evaluation,
		CapabilityProjection,
		ReasonMissingAltitude,
	)
}

func TestEvaluateRejectsProjectionWithoutRecentContinuity(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)

	item.Points[len(item.Points)-2].ObservedAt =
		now.Add(-3 * time.Minute)
	item.Points[len(item.Points)-1].ObservedAt =
		now.Add(-30 * time.Second)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireDeniedWithReason(
		t,
		evaluation,
		CapabilityProjection,
		ReasonInsufficientRecentContinuity,
	)

	for _, capability := range []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
	} {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}
}

func TestEvaluateAppliesCapabilitySpecificPointThresholds(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.PointCount = 1
	item.Points = item.Points[:1]

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireAllowed(
		t,
		evaluation,
		CapabilityTrafficMetrics,
	)

	for _, capability := range []Capability{
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonInsufficientPoints,
		)
	}
}

func TestEvaluateRejectsInvalidTimeRangeForEveryCapability(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.StartTime = now
	item.EndTime =
		now.Add(-time.Minute)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonInvalidTimeRange,
		)
	}
}

func TestEvaluateDoesNotMutateInputAndIsDeterministic(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)

	item.Points[0], item.Points[len(item.Points)-1] =
		item.Points[len(item.Points)-1],
		item.Points[0]

	original := deepCopyTrajectory(item)
	evaluator := NewDefault()

	first := evaluator.Evaluate(
		item,
		now,
	)
	second := evaluator.Evaluate(
		item,
		now,
	)

	if !reflect.DeepEqual(
		first,
		second,
	) {
		t.Fatalf(
			"expected deterministic evaluations, got %#v and %#v",
			first,
			second,
		)
	}

	if !reflect.DeepEqual(
		item,
		original,
	) {
		t.Fatal("expected evaluator not to mutate the trajectory")
	}
}

func TestEvaluateSupportsReliableSourceFlightIdentity(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.IdentityBasis =
		trajectory.FlightIdentityBasisSourceFlightID
	item.FlightID =
		"11111111-1111-1111-1111-111111111111"
	item.Callsign = ""

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}
}

func TestEvaluateRejectsSourceIdentityWithoutValidUUID(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.IdentityBasis =
		trajectory.FlightIdentityBasisSourceFlightID
	item.FlightID = "not-a-uuid"

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	requireAllowed(
		t,
		evaluation,
		CapabilityTrafficMetrics,
	)

	for _, capability := range []Capability{
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonIdentityNotReliable,
		)
	}
}

func TestEvaluateReportsDurationBoundaries(
	t *testing.T,
) {
	now := eligibilityTestTime()

	tooShort := healthyTrajectory(now)
	tooShort.StartTime =
		tooShort.EndTime.Add(-30 * time.Second)

	shortEvaluation := NewDefault().Evaluate(
		tooShort,
		now,
	)

	for _, capability := range []Capability{
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			shortEvaluation,
			capability,
			ReasonDurationTooShort,
		)
	}

	tooLong := healthyTrajectory(now)
	tooLong.StartTime =
		tooLong.EndTime.Add(-25 * time.Hour)

	longEvaluation := NewDefault().Evaluate(
		tooLong,
		now,
	)

	for _, capability := range []Capability{
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	} {
		requireDeniedWithReason(
			t,
			longEvaluation,
			capability,
			ReasonDurationTooLong,
		)
	}
}

func TestEvaluateReportsMissingAircraftIdentifierForEveryCapability(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.ICAO24 = "   "

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonMissingAircraftIdentifier,
		)
	}
}

func TestEvaluateSupportsCustomCallsignRequirement(
	t *testing.T,
) {
	now := eligibilityTestTime()
	config := DefaultConfig()
	config.TrafficMetrics.RequireCallsign = true

	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"expected valid custom config, got %v",
			err,
		)
	}

	item := healthyTrajectory(now)
	item.Callsign = ""

	evaluation := evaluator.Evaluate(
		item,
		now,
	)

	requireDeniedWithReason(
		t,
		evaluation,
		CapabilityTrafficMetrics,
		ReasonMissingCallsign,
	)
}

func TestEvaluateProducesStableReasonOrder(
	t *testing.T,
) {
	evaluation := NewDefault().Evaluate(
		trajectory.FlightTrajectory{},
		time.Time{},
	)

	decision := requireDecision(
		t,
		evaluation,
		CapabilityProjection,
	)

	expected := []ReasonCode{
		ReasonMissingAircraftIdentifier,
		ReasonInvalidTimeRange,
		ReasonInsufficientPoints,
		ReasonLowQualityScore,
		ReasonMissingIdentity,
		ReasonMissingAltitude,
		ReasonEvaluationTimeMissing,
		ReasonInsufficientRecentContinuity,
	}

	if !reflect.DeepEqual(
		decision.Reasons,
		expected,
	) {
		t.Fatalf(
			"expected stable reasons %v, got %v",
			expected,
			decision.Reasons,
		)
	}
}

func healthyTrajectory(
	now time.Time,
) trajectory.FlightTrajectory {
	endTime := now.Add(
		-30 * time.Second,
	)
	startTime := endTime.Add(
		-5 * time.Minute,
	)
	points := eligibilityPoints(
		endTime,
	)

	return trajectory.FlightTrajectory{
		IdentityKey: "flight-identity-" +
			strings.Repeat(
				"a",
				64,
			),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:          "ABC123",
		Callsign:        "AHY101",
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: 300,
		SegmentCount:    1,
		PointCount:      len(points),
		QualityScore:    0.90,
		SourceName:      "test",
		Points:          points,
	}
}

func eligibilityPoints(
	endTime time.Time,
) []trajectory.TrackPoint4D {
	result := make(
		[]trajectory.TrackPoint4D,
		0,
		6,
	)

	for index := 0; index < 6; index++ {
		result = append(
			result,
			trajectory.TrackPoint4D{
				ID:       fmtPointID(index),
				ICAO24:   "ABC123",
				Callsign: "AHY101",
				Latitude: 40.0 +
					float64(index)*0.01,
				Longitude: 49.0 +
					float64(index)*0.01,
				BarometricAltitudeM: 1000 +
					float64(index)*100,
				BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
				ObservedAt: endTime.Add(
					-time.Duration(
						5-index,
					) * time.Minute,
				),
				SourceName: "test",
			},
		)
	}

	return result
}

func fmtPointID(
	index int,
) string {
	return string(
		rune(
			'a' + index,
		),
	)
}

func eligibilityTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		13,
		13,
		0,
		0,
		0,
		time.UTC,
	)
}

func requireDecision(
	t *testing.T,
	evaluation Evaluation,
	capability Capability,
) Decision {
	t.Helper()

	decision, exists := evaluation.Decision(
		capability,
	)
	if !exists {
		t.Fatalf(
			"expected decision for %s",
			capability,
		)
	}

	return decision
}

func requireAllowed(
	t *testing.T,
	evaluation Evaluation,
	capability Capability,
) {
	t.Helper()

	decision := requireDecision(
		t,
		evaluation,
		capability,
	)
	if !decision.Allowed {
		t.Fatalf(
			"expected %s to be allowed, got reasons %v",
			capability,
			decision.Reasons,
		)
	}

	if !evaluation.Permissions.Allowed(
		capability,
	) {
		t.Fatalf(
			"expected %s permission flag to be true",
			capability,
		)
	}
}

func requireDeniedWithReason(
	t *testing.T,
	evaluation Evaluation,
	capability Capability,
	reason ReasonCode,
) {
	t.Helper()

	decision := requireDecision(
		t,
		evaluation,
		capability,
	)
	if decision.Allowed {
		t.Fatalf(
			"expected %s to be denied",
			capability,
		)
	}

	if !decision.HasReason(reason) {
		t.Fatalf(
			"expected %s denial reason %s, got %v",
			capability,
			reason,
			decision.Reasons,
		)
	}

	if evaluation.Permissions.Allowed(
		capability,
	) {
		t.Fatalf(
			"expected %s permission flag to be false",
			capability,
		)
	}
}

func deepCopyTrajectory(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	result := item
	result.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	result.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	result.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	return result
}

func TestEvaluateRejectsObservationBeyondFutureSkew(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.EndTime = now.Add(
		DefaultMaximumFutureObservationSkew +
			time.Second,
	)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireDeniedWithReason(
			t,
			evaluation,
			capability,
			ReasonFutureObservation,
		)
	}
}

func TestEvaluateAllowsObservationAtFutureSkewBoundary(
	t *testing.T,
) {
	now := eligibilityTestTime()
	item := healthyTrajectory(now)
	item.EndTime = now.Add(
		DefaultMaximumFutureObservationSkew,
	)

	evaluation := NewDefault().Evaluate(
		item,
		now,
	)

	for _, capability := range Capabilities() {
		requireAllowed(
			t,
			evaluation,
			capability,
		)
	}
}
