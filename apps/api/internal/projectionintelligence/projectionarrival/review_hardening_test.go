package projectionarrival

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestEstimateWithholdsAircraftMovingAwayFromDestination(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{-0.01, -0.02, -0.03},
	)

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	if result.Arrival != nil ||
		!hasReasonText(
			result.Limitations,
			string(unavailableSpeedOrDuration),
		) {
		t.Fatalf(
			"receding trajectory produced arrival: %#v",
			result.Arrival,
		)
	}
}

func TestEstimateWithholdsPhysicallyImpossibleGroundSpeed(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		10,
		[]float64{1, 2, 3},
	)

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	if result.Arrival != nil ||
		!hasReasonText(
			result.Limitations,
			string(unavailableSpeedOrDuration),
		) {
		t.Fatalf(
			"impossible ground speed produced arrival: %#v",
			result.Arrival,
		)
	}
}

func TestEstimateRequiresCurrentTrajectoryIdentifier(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{0.01, 0.02, 0.03},
	)
	request.CurrentTrajectory.ID = ""

	_, err := estimator.Estimate(request)
	if !errors.Is(err, ErrCurrentTrajectoryIDRequired) {
		t.Fatalf(
			"Estimate() error = %v, want %v",
			err,
			ErrCurrentTrajectoryIDRequired,
		)
	}
}

func TestArrivalFingerprintBindsCurrentEndpointEvidence(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{0.01, 0.02, 0.03},
	)

	first, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	request.CurrentTrajectory.Points[1].ID =
		"current-endpoint-reidentified"
	second, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf("second Estimate() error = %v", err)
	}

	if first.Provenance.InputFingerprint ==
		second.Provenance.InputFingerprint {
		t.Fatal(
			"current endpoint evidence identity did not change arrival fingerprint",
		)
	}
	if !hasInputReference(
		first.Provenance.Inputs,
		"current_trajectory_arrival_endpoint",
	) {
		t.Fatalf(
			"current trajectory input missing: %#v",
			first.Provenance.Inputs,
		)
	}
}

func TestRadiusCrossingUncertaintyUsesRadialClosingSpeed(
	t *testing.T,
) {
	config := validArrivalConfig()
	config.MinimumArrivalInterval = 2 * time.Second
	estimator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	start := arrivalTestAsOfTime()
	previous := positionSample{
		timeValue:              start.Add(time.Minute),
		latitude:               0.0099,
		longitude:              0,
		horizontalUncertaintyM: 100,
	}
	current := positionSample{
		timeValue:              start.Add(2 * time.Minute),
		latitude:               0,
		longitude:              0.0081,
		horizontalUncertaintyM: 100,
	}
	previousDistanceM := greatCircleDistanceM(
		previous.latitude,
		previous.longitude,
		0,
		0,
	)
	currentDistanceM := greatCircleDistanceM(
		current.latitude,
		current.longitude,
		0,
		0,
	)

	computation, valid := estimator.arrivalAtRadiusCrossing(
		previous,
		current,
		previousDistanceM,
		currentDistanceM,
		speedProfile{},
		false,
		validProjectionResult(start, []float64{0.01, 0.02}),
	)
	if !valid {
		t.Fatal("arrivalAtRadiusCrossing() returned invalid")
	}
	if computation.latestTime.Sub(computation.estimatedTime) <
		25*time.Second {
		t.Fatalf(
			"latest interval half-width = %s, want radial closing-speed uncertainty",
			computation.latestTime.Sub(computation.estimatedTime),
		)
	}
}

func TestExtrapolatedArrivalRejectsLatestTimeBeyondMaximumDuration(
	t *testing.T,
) {
	config := validArrivalConfig()
	config.MinimumArrivalInterval = 6 * time.Minute
	config.MaximumEstimatedArrivalDuration = 10 * time.Minute
	estimator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	asOfTime := arrivalTestAsOfTime()
	projection := validProjectionResult(
		asOfTime,
		[]float64{0.01, 0.02},
	)
	lastTime := projection.Horizon.EndTime
	samples := []positionSample{
		{timeValue: lastTime.Add(-time.Minute)},
		{timeValue: lastTime},
	}
	distances := []float64{6000, 5800}
	profile := speedProfile{
		sampleCount:            2,
		meanClosingSpeedMPS:    10,
		closingSpeedStdDevMPS:  0,
		minimumClosingSpeedMPS: 10,
		maximumClosingSpeedMPS: 10,
		maximumGroundSpeedMPS:  10,
	}

	_, valid := estimator.extrapolatedArrival(
		samples,
		distances,
		profile,
		true,
		projection,
	)
	if valid {
		t.Fatal(
			"minimum interval expanded LatestTime beyond maximum duration but arrival remained available",
		)
	}
}

func TestArrivalConfidenceReasonsReconstructFinalScore(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{0.01, 0.02, 0.03},
	)

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	if result.Arrival == nil {
		t.Fatal("arrival estimate is nil")
	}

	sum := 0.0
	for _, reason := range result.Arrival.Confidence.Reasons {
		sum += reason.Contribution
	}
	if math.Abs(sum-result.Arrival.Confidence.Score) > 1e-12 {
		t.Fatalf(
			"reason contribution sum = %.17g, score = %.17g",
			sum,
			result.Arrival.Confidence.Score,
		)
	}
}

func TestNewNormalizesMaximumGroundSpeed(
	t *testing.T,
) {
	config := validArrivalConfig()
	config.MaximumGroundSpeedMPS = 0

	estimator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if estimator.config.MaximumGroundSpeedMPS !=
		defaultMaximumGroundSpeedMPS {
		t.Fatalf(
			"maximum ground speed = %f, want %f",
			estimator.config.MaximumGroundSpeedMPS,
			defaultMaximumGroundSpeedMPS,
		)
	}
}

func hasInputReference(
	items []projectioncontract.InputReference,
	name string,
) bool {
	for _, item := range items {
		if item.Name == name {
			return true
		}
	}
	return false
}
