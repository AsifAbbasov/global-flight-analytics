package projectionevaluation

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestEvaluationFingerprintBindsProjectionOutput(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	leftRequest := evaluationTestRequest(true)
	rightRequest := evaluationTestRequest(true)
	rightRequest.Projection.Points[0].Position.Longitude += 0.001
	left, err := evaluator.Evaluate(leftRequest)
	if err != nil {
		t.Fatal(err)
	}
	right, err := evaluator.Evaluate(rightRequest)
	if err != nil {
		t.Fatal(err)
	}
	if left.ProjectionInputFingerprint != right.ProjectionInputFingerprint {
		t.Fatal("fixture changed upstream projection input fingerprint")
	}
	if left.ProjectionSnapshotFingerprint == right.ProjectionSnapshotFingerprint ||
		left.EvaluationInputFingerprint == right.EvaluationInputFingerprint {
		t.Fatal("projection output mutation did not change evaluation lineage")
	}
}

func TestEvaluationFingerprintBindsAltitudeStatuses(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	leftRequest := evaluationTestRequest(false)
	rightRequest := evaluationTestRequest(false)
	rightRequest.ActualTrajectory.Points[0].GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable
	rightRequest.ActualTrajectory.Points[0].BarometricAltitudeStatus = flightstate.AltitudeStatusUnavailable
	left, err := evaluator.Evaluate(leftRequest)
	if err != nil {
		t.Fatal(err)
	}
	right, err := evaluator.Evaluate(rightRequest)
	if err != nil {
		t.Fatal(err)
	}
	if left.TruthSnapshotFingerprint == right.TruthSnapshotFingerprint ||
		left.EvaluationInputFingerprint == right.EvaluationInputFingerprint {
		t.Fatal("altitude status mutation did not change truth lineage")
	}
	if left.Position.AltitudeEvaluatedPointCount == right.Position.AltitudeEvaluatedPointCount {
		t.Fatal("altitude status mutation did not change altitude evaluation")
	}
}

func TestEvaluationIsOrderIndependentForIdenticalTruthDuplicates(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	leftRequest := evaluationTestRequest(false)
	duplicate := leftRequest.ActualTrajectory.Points[0]
	duplicate.ID = "actual-point-duplicate"
	leftRequest.ActualTrajectory.Points = append(leftRequest.ActualTrajectory.Points, duplicate)
	leftRequest.TruthAvailability = append(leftRequest.TruthAvailability, TruthAvailability{
		PointID:     duplicate.ID,
		SourceName:  "actual-truth-ingestion",
		AvailableAt: leftRequest.TruthAvailability[0].AvailableAt,
	})
	rightRequest := leftRequest
	rightRequest.ActualTrajectory = leftRequest.ActualTrajectory
	rightRequest.ActualTrajectory.Points = append([]trajectory.TrackPoint4D(nil), leftRequest.ActualTrajectory.Points...)
	rightRequest.ActualTrajectory.Points[0], rightRequest.ActualTrajectory.Points[len(rightRequest.ActualTrajectory.Points)-1] =
		rightRequest.ActualTrajectory.Points[len(rightRequest.ActualTrajectory.Points)-1], rightRequest.ActualTrajectory.Points[0]
	left, err := evaluator.Evaluate(leftRequest)
	if err != nil {
		t.Fatal(err)
	}
	right, err := evaluator.Evaluate(rightRequest)
	if err != nil {
		t.Fatal(err)
	}
	if left.EvaluationInputFingerprint != right.EvaluationInputFingerprint ||
		left.Position != right.Position {
		t.Fatal("identical duplicate truth order changed evaluation")
	}
}

func TestEvaluatePublishesEndpointLeadTimeConfidenceAndPolicy(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	result, err := evaluator.Evaluate(evaluationTestRequest(true))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Endpoint.Available || result.Endpoint.ForecastTime != result.ProjectionHorizonEndTime {
		t.Fatalf("endpoint metrics unavailable: %#v", result.Endpoint)
	}
	if len(result.LeadTimes) != 3 || result.Confidence.EvaluatedPointCount != 3 {
		t.Fatalf("lead-time or confidence metrics missing: lead=%#v confidence=%#v", result.LeadTimes, result.Confidence)
	}
	if result.Policy.Version != EvaluationPolicyVersion || !fingerprintPattern.MatchString(result.Policy.InputFingerprint) {
		t.Fatalf("policy snapshot missing: %#v", result.Policy)
	}
}

func TestEvaluateRejectsInvalidActualArrivalICAO(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	request := evaluationTestRequest(true)
	request.ActualArrival.AirportICAOCode = "!@#$"
	_, err := evaluator.Evaluate(request)
	if !errors.Is(err, ErrActualArrivalInvalid) {
		t.Fatalf("error = %v", err)
	}
}

func TestEvaluateRejectsImplausibleTruthInterpolation(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	request := evaluationTestRequest(false)
	asOf := request.Projection.Horizon.AsOfTime
	request.Projection.Points = request.Projection.Points[:1]
	request.Projection.Horizon.EndTime = asOf.Add(time.Minute)
	request.ActualTrajectory.Points = []trajectory.TrackPoint4D{
		{ID: "left", Latitude: 0, Longitude: 0, ObservedAt: asOf.Add(30 * time.Second), SourceName: "truth"},
		{ID: "right", Latitude: 0, Longitude: 10, ObservedAt: asOf.Add(90 * time.Second), SourceName: "truth"},
	}
	request.TruthAvailability = []TruthAvailability{
		{PointID: "left", SourceName: "ingest", AvailableAt: asOf.Add(31 * time.Second)},
		{PointID: "right", SourceName: "ingest", AvailableAt: asOf.Add(91 * time.Second)},
	}
	result, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != StatusUnavailable || result.Position.EvaluatedPointCount != 0 ||
		!hasEvaluationNotice(result.Limitations, "implausible_truth_interpolation_rejected") {
		t.Fatalf("implausible movement became truth: %#v", result)
	}
}

func TestResultValidationRejectsTamperedDerivedMetrics(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	result, err := evaluator.Evaluate(evaluationTestRequest(true))
	if err != nil {
		t.Fatal(err)
	}

	tamperedPoint := result.Clone()
	tamperedPoint.Points[0].HorizontalErrorM++
	if tamperedPoint.Validate() == nil {
		t.Fatal("tampered point error passed validation")
	}

	tamperedCoverage := result.Clone()
	tamperedCoverage.Position.CoverageRatio = 0.5
	if tamperedCoverage.Validate() == nil {
		t.Fatal("tampered coverage passed validation")
	}

	tamperedArrival := result.Clone()
	tamperedArrival.Arrival.IntervalWidthSeconds++
	if tamperedArrival.Validate() == nil {
		t.Fatal("tampered arrival interval passed validation")
	}
}

func TestMedianUsesAverageOfEvenMiddleValues(t *testing.T) {
	if got := median([]float64{1, 2, 100, 200}); got != 51 {
		t.Fatalf("median = %f, want 51", got)
	}
}

func TestAggregateSeparatesDecisionClassAndPolicy(t *testing.T) {
	firstEvaluator := newEvaluationEvaluator(t)
	first, err := firstEvaluator.Evaluate(evaluationTestRequest(false))
	if err != nil {
		t.Fatal(err)
	}

	classRequest := evaluationTestRequest(false)
	classRequest.Projection.Method.DecisionClass = projectioncontract.DecisionClassPhysicsDerived
	classRequest.Projection.Provenance.InputFingerprint = "sha256:" + strings.Repeat("c", 64)
	second, err := firstEvaluator.Evaluate(classRequest)
	if err != nil {
		t.Fatal(err)
	}

	config := validEvaluationConfig()
	config.LeadTimeBucketSize = 2 * time.Minute
	policyEvaluator, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	policyRequest := evaluationTestRequest(false)
	policyRequest.Projection.Provenance.InputFingerprint = "sha256:" + strings.Repeat("d", 64)
	third, err := policyEvaluator.Evaluate(policyRequest)
	if err != nil {
		t.Fatal(err)
	}

	aggregate, err := Aggregate([]Result{third, second, first}, evaluationTestAsOfTime().Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if aggregate.MethodCount != 3 {
		t.Fatalf("method groups = %d, want 3: %#v", aggregate.MethodCount, aggregate.Methods)
	}
}

func TestAggregateExcludesUnavailableEvaluationFromAccuracyMetrics(t *testing.T) {
	config := validEvaluationConfig()
	config.MinimumEvaluatedPointCount = 2
	evaluator, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	complete, err := evaluator.Evaluate(evaluationTestRequest(false))
	if err != nil {
		t.Fatal(err)
	}
	unavailableRequest := evaluationTestRequest(false)
	unavailableRequest.ActualTrajectory.Points = unavailableRequest.ActualTrajectory.Points[:1]
	unavailableRequest.TruthAvailability = unavailableRequest.TruthAvailability[:1]
	unavailableRequest.ActualTrajectory.Points[0].Longitude = 50
	unavailable, err := evaluator.Evaluate(unavailableRequest)
	if err != nil {
		t.Fatal(err)
	}
	if unavailable.Status != StatusUnavailable || len(unavailable.Points) != 1 {
		t.Fatalf("fixture not unavailable: %#v", unavailable)
	}

	aggregate, err := Aggregate([]Result{complete, unavailable}, evaluationTestAsOfTime().Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	summary := aggregate.Methods[0]
	if summary.UnavailableEvaluationCount != 1 || summary.AccuracyEligibleEvaluationCount != 1 ||
		summary.EvaluatedPointCount != complete.Position.EvaluatedPointCount ||
		!almostEqual(summary.MeanHorizontalErrorM, complete.Position.MeanHorizontalErrorM) {
		t.Fatalf("unavailable evaluation contaminated accuracy: %#v", summary)
	}
}

func TestAggregatePublishesArrivalSelectivePredictionAccounting(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	matched, err := evaluator.Evaluate(evaluationTestRequest(true))
	if err != nil {
		t.Fatal(err)
	}
	missingRequest := evaluationTestRequest(true)
	missingRequest.Projection.Arrival = nil
	missing, err := evaluator.Evaluate(missingRequest)
	if err != nil {
		t.Fatal(err)
	}
	mismatchRequest := evaluationTestRequest(true)
	mismatchRequest.ActualArrival.AirportICAOCode = "CCCC"
	mismatch, err := evaluator.Evaluate(mismatchRequest)
	if err != nil {
		t.Fatal(err)
	}

	aggregate, err := Aggregate([]Result{matched, missing, mismatch}, evaluationTestAsOfTime().Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	summary := aggregate.Methods[0]
	if summary.ActualArrivalTruthCount != 3 || summary.ArrivalPredictionCount != 2 ||
		summary.MatchedArrivalCount != 1 || summary.MissingArrivalPredictionCount != 1 ||
		summary.ArrivalAirportMismatchCount != 1 || summary.ArrivalEvaluationCount != 1 ||
		math.Abs(summary.ArrivalPredictionRecall-2.0/3.0) > 1e-12 || summary.ArrivalAirportAccuracy != 0.5 {
		t.Fatalf("selective arrival accounting is wrong: %#v", summary)
	}
}

func TestAggregateFingerprintExcludesGeneratedAt(t *testing.T) {
	evaluator := newEvaluationEvaluator(t)
	result, err := evaluator.Evaluate(evaluationTestRequest(false))
	if err != nil {
		t.Fatal(err)
	}
	left, err := Aggregate([]Result{result}, evaluationTestAsOfTime().Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	right, err := Aggregate([]Result{result}, evaluationTestAsOfTime().Add(20*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if left.InputFingerprint != right.InputFingerprint {
		t.Fatal("aggregate input fingerprint depends on GeneratedAt")
	}
	if left.GeneratedAt == right.GeneratedAt {
		t.Fatal("fixture generated times did not differ")
	}
}
