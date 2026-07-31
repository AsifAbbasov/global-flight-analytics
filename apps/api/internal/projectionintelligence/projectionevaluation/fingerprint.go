package projectionevaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const fingerprintPrefix = "sha256:"

func evaluationPolicyFingerprint(policy EvaluationPolicy) string {
	digest := sha256.New()
	writeFingerprintString(digest, policy.Version)
	writeFingerprintDuration(digest, policy.MaximumInterpolationGap)
	writeFingerprintFloat(digest, policy.MaximumTruthGroundSpeedMPS)
	writeFingerprintFloat(digest, policy.MaximumTruthVerticalRateMPS)
	writeFingerprintInt(digest, policy.MinimumEvaluatedPointCount)
	writeFingerprintFloat(digest, policy.MaximumHorizontalErrorM)
	writeFingerprintFloat(digest, policy.MaximumAltitudeErrorM)
	writeFingerprintDuration(digest, policy.LeadTimeBucketSize)
	return sumFingerprint(digest)
}

func projectionSnapshotFingerprint(projection projectioncontract.Result) string {
	digest := sha256.New()
	writeFingerprintString(digest, ProjectionSnapshotVersion)
	writeFingerprintString(digest, string(projection.SchemaVersion))
	writeFingerprintString(digest, string(projection.Status))
	writeFingerprintString(digest, projection.TrajectoryID)
	writeFingerprintString(digest, projection.FlightID)
	writeFingerprintString(digest, projection.AircraftID)
	writeFingerprintString(digest, projection.ICAO24)
	writeFingerprintString(digest, projection.Callsign)
	writeFingerprintString(digest, projection.Method.Name)
	writeFingerprintString(digest, projection.Method.Version)
	writeFingerprintString(digest, string(projection.Method.DecisionClass))
	writeFingerprintTime(digest, projection.Horizon.AsOfTime)
	writeFingerprintTime(digest, projection.Horizon.EndTime)
	writeFingerprintDuration(digest, projection.Horizon.Step)
	writeFingerprintInt(digest, len(projection.Points))
	for _, point := range projection.Points {
		writeFingerprintInt(digest, point.Sequence)
		writeFingerprintTime(digest, point.ForecastTime)
		writeFingerprintFloat(digest, point.Position.Latitude)
		writeFingerprintFloat(digest, point.Position.Longitude)
		writeOptionalFloat(digest, point.Position.AltitudeM)
		writeFingerprintFloat(digest, point.Uncertainty.HorizontalRadiusM)
		writeOptionalFloat(digest, point.Uncertainty.VerticalRadiusM)
		writeConfidence(digest, point.Confidence)
	}
	if projection.Arrival == nil {
		writeFingerprintString(digest, "arrival:nil")
	} else {
		writeFingerprintString(digest, "arrival:present")
		writeFingerprintString(digest, projection.Arrival.AirportICAOCode)
		writeFingerprintTime(digest, projection.Arrival.EarliestTime)
		writeFingerprintTime(digest, projection.Arrival.EstimatedTime)
		writeFingerprintTime(digest, projection.Arrival.LatestTime)
		writeConfidence(digest, projection.Arrival.Confidence)
		writeFingerprintInt(digest, len(projection.Arrival.Limitations))
		for _, limitation := range projection.Arrival.Limitations {
			writeFingerprintString(digest, limitation.Code)
			writeFingerprintString(digest, limitation.Message)
			writeFingerprintString(digest, limitation.Scope)
		}
	}
	writeConfidence(digest, projection.Confidence)
	writeFingerprintInt(digest, len(projection.Limitations))
	for _, limitation := range projection.Limitations {
		writeFingerprintString(digest, limitation.Code)
		writeFingerprintString(digest, limitation.Message)
		writeFingerprintString(digest, limitation.Scope)
	}
	writeFingerprintInt(digest, len(projection.Explanations))
	for _, explanation := range projection.Explanations {
		writeFingerprintString(digest, explanation.Code)
		writeFingerprintString(digest, explanation.Message)
	}
	writeFingerprintString(digest, string(projection.ScopeGuard))
	writeFingerprintString(digest, projection.Provenance.InputFingerprint)
	writeFingerprintInt(digest, len(projection.Provenance.Inputs))
	for _, input := range projection.Provenance.Inputs {
		writeFingerprintString(digest, input.Name)
		writeFingerprintString(digest, string(input.Classification))
		writeFingerprintString(digest, input.SourceName)
		writeFingerprintTime(digest, input.ObservedAt)
		writeFingerprintTime(digest, input.RetrievedAt)
		writeFingerprintString(digest, input.Limitation)
	}
	writeFingerprintTime(digest, projection.Provenance.LatestInputObservedAt)
	writeFingerprintTime(digest, projection.GeneratedAt)
	return sumFingerprint(digest)
}

func truthSnapshotFingerprint(
	trajectoryID string,
	truth normalizedTruth,
) string {
	digest := sha256.New()
	writeFingerprintString(digest, TruthSnapshotVersion)
	writeFingerprintString(digest, trajectoryID)
	writeFingerprintString(digest, TruthKnowledgeCutoffMode)
	writeFingerprintInt(digest, truth.excludedAfterObservationCutoff)
	writeFingerprintInt(digest, truth.excludedAfterAvailabilityCutoff)
	writeFingerprintInt(digest, len(truth.points))
	for _, item := range truth.points {
		point := item.point
		writeFingerprintString(digest, point.ID)
		writeFingerprintTime(digest, point.ObservedAt)
		writeFingerprintTime(digest, item.availableAt)
		writeFingerprintString(digest, item.evidenceSource)
		writeFingerprintFloat(digest, point.Latitude)
		writeFingerprintFloat(digest, point.Longitude)
		writeFingerprintFloat(digest, point.GeometricAltitudeM)
		writeFingerprintString(digest, string(point.GeometricAltitudeStatus))
		writeFingerprintFloat(digest, point.BarometricAltitudeM)
		writeFingerprintString(digest, string(point.BarometricAltitudeStatus))
		writeFingerprintString(digest, point.SourceName)
	}
	return sumFingerprint(digest)
}

func evaluationFingerprint(
	projectionFingerprint string,
	truthFingerprint string,
	actualArrival *ActualArrival,
	evaluatedAt time.Time,
	policy EvaluationPolicy,
) string {
	digest := sha256.New()
	writeFingerprintString(digest, FingerprintVersion)
	writeFingerprintString(digest, projectionFingerprint)
	writeFingerprintString(digest, truthFingerprint)
	writeFingerprintString(digest, policy.InputFingerprint)
	writeFingerprintTime(digest, evaluatedAt)
	if actualArrival == nil {
		writeFingerprintString(digest, "actual-arrival:nil")
	} else {
		writeFingerprintString(digest, "actual-arrival:present")
		writeFingerprintString(digest, actualArrival.AirportICAOCode)
		writeFingerprintTime(digest, actualArrival.BoundaryTime)
		writeFingerprintString(digest, actualArrival.SourceName)
		writeFingerprintTime(digest, actualArrival.ObservedAt)
		writeFingerprintTime(digest, actualArrival.AvailableAt)
	}
	return sumFingerprint(digest)
}

func aggregateFingerprint(results []Result) string {
	digest := sha256.New()
	writeFingerprintString(digest, AggregateFingerprintVersion)
	fingerprints := make([]string, 0, len(results))
	for _, result := range results {
		fingerprints = append(fingerprints, result.EvaluationInputFingerprint)
	}
	sort.Strings(fingerprints)
	writeFingerprintInt(digest, len(fingerprints))
	for _, fingerprint := range fingerprints {
		writeFingerprintString(digest, fingerprint)
	}
	return sumFingerprint(digest)
}

func writeConfidence(digest hash.Hash, confidence projectioncontract.Confidence) {
	writeFingerprintFloat(digest, confidence.Score)
	writeFingerprintString(digest, string(confidence.Level))
	writeFingerprintInt(digest, len(confidence.Reasons))
	for _, reason := range confidence.Reasons {
		writeFingerprintString(digest, reason.Code)
		writeFingerprintString(digest, reason.Message)
		writeFingerprintFloat(digest, reason.Contribution)
	}
}

func writeOptionalFloat(digest hash.Hash, value *float64) {
	if value == nil {
		writeFingerprintString(digest, "nil")
		return
	}
	writeFingerprintString(digest, "present")
	writeFingerprintFloat(digest, *value)
}

func sumFingerprint(digest hash.Hash) string {
	return fingerprintPrefix + hex.EncodeToString(digest.Sum(nil))
}

func writeFingerprintString(digest hash.Hash, value string) {
	_, _ = fmt.Fprintf(digest, "%d:%s|", len(value), value)
}

func writeFingerprintInt(digest hash.Hash, value int) {
	_, _ = fmt.Fprintf(digest, "%d|", value)
}

func writeFingerprintFloat(digest hash.Hash, value float64) {
	_, _ = fmt.Fprintf(digest, "%.17g|", value)
}

func writeFingerprintTime(digest hash.Hash, value time.Time) {
	writeFingerprintString(digest, value.UTC().Format(time.RFC3339Nano))
}

func writeFingerprintDuration(digest hash.Hash, value time.Duration) {
	_, _ = fmt.Fprintf(digest, "%d|", value.Nanoseconds())
}
