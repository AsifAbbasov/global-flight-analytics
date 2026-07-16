package forecaststability

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

type canonicalProjection struct {
	SchemaVersion projectioncontract.SchemaVersion
	Status        projectioncontract.ResultStatus
	TrajectoryID  string
	FlightID      string
	AircraftID    string
	ICAO24        string
	Callsign      string
	Method        projectioncontract.Method
	Horizon       projectioncontract.Horizon
	Points        []projectioncontract.ProjectionPoint
	Arrival       *projectioncontract.ArrivalEstimate
	Confidence    projectioncontract.Confidence
	Limitations   []projectioncontract.Limitation
	Explanations  []projectioncontract.Explanation
	ScopeGuard    projectioncontract.ScopeGuard
	Provenance    projectioncontract.Provenance
}

func projectionOutputFingerprint(result projectioncontract.Result) string {
	normalized := result.Clone()
	normalizeProjection(&normalized)
	payload := canonicalProjection{
		SchemaVersion: normalized.SchemaVersion,
		Status:        normalized.Status,
		TrajectoryID:  normalized.TrajectoryID,
		FlightID:      normalized.FlightID,
		AircraftID:    normalized.AircraftID,
		ICAO24:        normalized.ICAO24,
		Callsign:      normalized.Callsign,
		Method:        normalized.Method,
		Horizon:       normalized.Horizon,
		Points:        normalized.Points,
		Arrival:       normalized.Arrival,
		Confidence:    normalized.Confidence,
		Limitations:   normalized.Limitations,
		Explanations:  normalized.Explanations,
		ScopeGuard:    normalized.ScopeGuard,
		Provenance:    normalized.Provenance,
	}
	return hashValue(payload)
}

func decisionFingerprint(
	projection projectioncontract.Result,
	policyVersion string,
	implementationVersion string,
) string {
	return hashValue(struct {
		TrajectoryID          string
		SchemaVersion         projectioncontract.SchemaVersion
		Method                projectioncontract.Method
		Horizon               projectioncontract.Horizon
		PolicyVersion         string
		ImplementationVersion string
		InputFingerprint      string
		OutputFingerprint     string
	}{
		TrajectoryID:          strings.TrimSpace(projection.TrajectoryID),
		SchemaVersion:         projection.SchemaVersion,
		Method:                projection.Method,
		Horizon:               normalizedHorizon(projection.Horizon),
		PolicyVersion:         strings.TrimSpace(policyVersion),
		ImplementationVersion: strings.TrimSpace(implementationVersion),
		InputFingerprint:      strings.TrimSpace(projection.Provenance.InputFingerprint),
		OutputFingerprint:     projectionOutputFingerprint(projection),
	})
}

func forecastVersionID(
	ordinal int,
	trajectoryID string,
	parentVersionID string,
	decisionFingerprintValue string,
) string {
	digest := hashValue(struct {
		Ordinal             int
		TrajectoryID        string
		ParentVersionID     string
		DecisionFingerprint string
	}{
		Ordinal:             ordinal,
		TrajectoryID:        strings.TrimSpace(trajectoryID),
		ParentVersionID:     strings.TrimSpace(parentVersionID),
		DecisionFingerprint: decisionFingerprintValue,
	})
	return "forecast-version-" + strings.TrimPrefix(digest, "sha256:")[:32]
}

func stabilityInputFingerprint(result StabilityResult, policy StabilityPolicy) string {
	copy := result.Clone()
	copy.Provenance.InputFingerprint = ""
	return hashValue(struct {
		Result StabilityResult
		Policy StabilityPolicy
	}{Result: copy, Policy: policy})
}

func hashValue(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("marshal deterministic fingerprint payload: %v", err))
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func normalizeProjection(result *projectioncontract.Result) {
	result.TrajectoryID = strings.TrimSpace(result.TrajectoryID)
	result.FlightID = strings.TrimSpace(result.FlightID)
	result.AircraftID = strings.TrimSpace(result.AircraftID)
	result.ICAO24 = strings.ToUpper(strings.TrimSpace(result.ICAO24))
	result.Callsign = strings.ToUpper(strings.TrimSpace(result.Callsign))
	result.Method.Name = strings.TrimSpace(result.Method.Name)
	result.Method.Version = strings.TrimSpace(result.Method.Version)
	result.Horizon = normalizedHorizon(result.Horizon)
	result.GeneratedAt = result.GeneratedAt.UTC()
	for index := range result.Points {
		result.Points[index].ForecastTime = result.Points[index].ForecastTime.UTC()
		sortConfidenceReasons(result.Points[index].Confidence.Reasons)
	}
	if result.Arrival != nil {
		result.Arrival.AirportICAOCode = strings.ToUpper(strings.TrimSpace(result.Arrival.AirportICAOCode))
		result.Arrival.EarliestTime = result.Arrival.EarliestTime.UTC()
		result.Arrival.EstimatedTime = result.Arrival.EstimatedTime.UTC()
		result.Arrival.LatestTime = result.Arrival.LatestTime.UTC()
		sortConfidenceReasons(result.Arrival.Confidence.Reasons)
		sort.Slice(result.Arrival.Limitations, func(left, right int) bool {
			return limitationKey(result.Arrival.Limitations[left]) < limitationKey(result.Arrival.Limitations[right])
		})
	}
	sortConfidenceReasons(result.Confidence.Reasons)
	sort.Slice(result.Limitations, func(left, right int) bool {
		return limitationKey(result.Limitations[left]) < limitationKey(result.Limitations[right])
	})
	sort.Slice(result.Explanations, func(left, right int) bool {
		return explanationKey(result.Explanations[left]) < explanationKey(result.Explanations[right])
	})
	sort.Slice(result.Provenance.Inputs, func(left, right int) bool {
		return inputReferenceKey(result.Provenance.Inputs[left]) < inputReferenceKey(result.Provenance.Inputs[right])
	})
	for index := range result.Provenance.Inputs {
		result.Provenance.Inputs[index].Name = strings.TrimSpace(result.Provenance.Inputs[index].Name)
		result.Provenance.Inputs[index].SourceName = strings.TrimSpace(result.Provenance.Inputs[index].SourceName)
		result.Provenance.Inputs[index].ObservedAt = result.Provenance.Inputs[index].ObservedAt.UTC()
		result.Provenance.Inputs[index].RetrievedAt = result.Provenance.Inputs[index].RetrievedAt.UTC()
		result.Provenance.Inputs[index].Limitation = strings.TrimSpace(result.Provenance.Inputs[index].Limitation)
	}
	result.Provenance.InputFingerprint = strings.TrimSpace(result.Provenance.InputFingerprint)
	result.Provenance.LatestInputObservedAt = result.Provenance.LatestInputObservedAt.UTC()
}

func normalizedHorizon(horizon projectioncontract.Horizon) projectioncontract.Horizon {
	horizon.AsOfTime = horizon.AsOfTime.UTC()
	horizon.EndTime = horizon.EndTime.UTC()
	return horizon
}

func sortConfidenceReasons(reasons []projectioncontract.ConfidenceReason) {
	sort.Slice(reasons, func(left, right int) bool {
		leftKey := reasons[left].Code + "|" + reasons[left].Message + "|" + fmt.Sprintf("%.12f", reasons[left].Contribution)
		rightKey := reasons[right].Code + "|" + reasons[right].Message + "|" + fmt.Sprintf("%.12f", reasons[right].Contribution)
		return leftKey < rightKey
	})
}

func limitationKey(item projectioncontract.Limitation) string {
	return item.Code + "|" + item.Scope + "|" + item.Message
}

func explanationKey(item projectioncontract.Explanation) string {
	return item.Code + "|" + item.Message
}

func inputReferenceKey(item projectioncontract.InputReference) string {
	return item.Name + "|" + string(item.Classification) + "|" + item.SourceName + "|" + item.ObservedAt.UTC().Format(time.RFC3339Nano) + "|" + item.RetrievedAt.UTC().Format(time.RFC3339Nano)
}
