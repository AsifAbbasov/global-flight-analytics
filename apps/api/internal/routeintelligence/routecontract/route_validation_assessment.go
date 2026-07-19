package routecontract

import (
	"fmt"
	"math"

	"strings"
)

func validateConfidence(
	confidence Confidence,
	fieldPrefix string,
	expectedEvidenceCount int,
	collector *validationCollector,
) {
	if !finiteRatio(confidence.Score) {
		collector.add(
			ValidationSeverityError,
			"confidence_score_invalid",
			fieldPrefix+".score",
			"Confidence score must be finite and between zero and one.",
		)
	} else if confidence.Level !=
		ConfidenceLevelForScore(confidence.Score) {
		collector.add(
			ValidationSeverityError,
			"confidence_level_mismatch",
			fieldPrefix+".level",
			"Confidence level does not match the score thresholds.",
		)
	}
	if confidence.EvidenceCount !=
		expectedEvidenceCount {
		collector.add(
			ValidationSeverityError,
			"confidence_evidence_count_mismatch",
			fieldPrefix+".evidence_count",
			fmt.Sprintf(
				"Confidence evidence count must equal %d.",
				expectedEvidenceCount,
			),
		)
	}

	seenCodes := make(map[string]struct{})
	for index, reason := range confidence.Reasons {
		itemPrefix := fmt.Sprintf(
			"%s.reasons[%d]",
			fieldPrefix,
			index,
		)
		code := strings.TrimSpace(reason.Code)
		if code == "" {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_code_required",
				itemPrefix+".code",
				"Confidence reason code is required.",
			)
		} else if _, exists := seenCodes[code]; exists {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_code_duplicate",
				itemPrefix+".code",
				"Confidence reason codes must be unique.",
			)
		} else {
			seenCodes[code] = struct{}{}
		}
		if strings.TrimSpace(reason.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_message_required",
				itemPrefix+".message",
				"Confidence reason message is required.",
			)
		}
		if math.IsNaN(reason.Contribution) ||
			math.IsInf(reason.Contribution, 0) ||
			reason.Contribution < -1 ||
			reason.Contribution > 1 {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_contribution_invalid",
				itemPrefix+".contribution",
				"Confidence reason contribution must be finite and between -1 and 1.",
			)
		}
	}
}

func validateSummary(
	result Result,
	collector *validationCollector,
) {
	if !finiteNonNegative(
		result.Summary.GreatCircleDistanceKM,
	) {
		collector.add(
			ValidationSeverityError,
			"route_distance_invalid",
			"summary.great_circle_distance_km",
			"Great-circle route distance must be finite and non-negative.",
		)
	}

	if result.Origin == nil ||
		result.Destination == nil {
		if result.Summary.SameAirport {
			collector.add(
				ValidationSeverityError,
				"same_airport_without_complete_route",
				"summary.same_airport",
				"Same-airport status requires both route endpoints.",
			)
		}
		if result.Summary.GreatCircleDistanceKM != 0 {
			collector.add(
				ValidationSeverityWarning,
				"route_distance_without_complete_route",
				"summary.great_circle_distance_km",
				"Route distance is normally zero until both endpoints are resolved.",
			)
		}

		return
	}

	expectedSameAirport := strings.EqualFold(
		result.Origin.Airport.ICAOCode,
		result.Destination.Airport.ICAOCode,
	)
	if result.Summary.SameAirport !=
		expectedSameAirport {
		collector.add(
			ValidationSeverityError,
			"same_airport_mismatch",
			"summary.same_airport",
			"Same-airport status must match the endpoint ICAO codes.",
		)
	}
}

func validateLimitations(
	items []Limitation,
	fieldPrefix string,
	collector *validationCollector,
) {
	seenCodes := make(map[string]struct{})
	for index, item := range items {
		itemPrefix := fmt.Sprintf(
			"%s[%d]",
			fieldPrefix,
			index,
		)
		code := strings.TrimSpace(item.Code)
		if code == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_code_required",
				itemPrefix+".code",
				"Limitation code is required.",
			)
		} else if _, exists := seenCodes[code]; exists {
			collector.add(
				ValidationSeverityError,
				"limitation_code_duplicate",
				itemPrefix+".code",
				"Limitation codes must be unique within their scope.",
			)
		} else {
			seenCodes[code] = struct{}{}
		}
		if strings.TrimSpace(item.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_message_required",
				itemPrefix+".message",
				"Limitation message is required.",
			)
		}
		if strings.TrimSpace(item.Scope) == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_scope_required",
				itemPrefix+".scope",
				"Limitation scope is required.",
			)
		}
	}
}
