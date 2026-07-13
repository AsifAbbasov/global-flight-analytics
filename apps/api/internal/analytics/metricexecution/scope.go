package metricexecution

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func buildScopeSummary(
	result scopeguard.FilterResult,
	capability trajectoryeligibility.Capability,
	inputCount int,
) ScopeSummary {
	counts := make(
		map[trajectoryeligibility.ReasonCode]int,
	)

	for _, denied := range result.Denied {
		for _, reason := range denied.Decision.Reasons {
			counts[reason]++
		}
	}

	reasons := make(
		[]ReasonCount,
		0,
		len(counts),
	)

	for reason, count := range counts {
		reasons = append(
			reasons,
			ReasonCount{
				Reason: reason,
				Count:  count,
			},
		)
	}

	sort.SliceStable(
		reasons,
		func(left int, right int) bool {
			return reasons[left].Reason < reasons[right].Reason
		},
	)

	return ScopeSummary{
		Capability:   capability,
		InputCount:   inputCount,
		AllowedCount: result.AllowedCount(),
		DeniedCount:  result.DeniedCount(),
		Reasons:      reasons,
		EvaluatedAt:  result.EvaluatedAt.UTC(),
	}
}

func aggregateDeniedDecision(
	result scopeguard.FilterResult,
	capability trajectoryeligibility.Capability,
) (scopeguard.Decision, error) {
	reasonSet := make(
		map[trajectoryeligibility.ReasonCode]struct{},
	)

	for _, denied := range result.Denied {
		for _, reason := range denied.Decision.Reasons {
			reasonSet[reason] = struct{}{}
		}
	}

	if len(reasonSet) == 0 {
		return scopeguard.Decision{},
			ErrAggregateDenialReasonsMissing
	}

	reasons := make(
		[]trajectoryeligibility.ReasonCode,
		0,
		len(reasonSet),
	)

	for reason := range reasonSet {
		reasons = append(reasons, reason)
	}

	sort.SliceStable(
		reasons,
		func(left int, right int) bool {
			return reasons[left] < reasons[right]
		},
	)

	return scopeguard.Decision{
		Capability:  capability,
		Allowed:     false,
		Reasons:     reasons,
		EvaluatedAt: result.EvaluatedAt.UTC(),
	}, nil
}

func uniqueTrajectories(
	items []trajectory.FlightTrajectory,
) ([]trajectory.FlightTrajectory, int) {
	result := make(
		[]trajectory.FlightTrajectory,
		0,
		len(items),
	)
	seen := make(
		map[string]struct{},
		len(items),
	)
	removed := 0

	for index, item := range items {
		key, stable := trajectoryContributorKey(item)
		if !stable {
			key = fmt.Sprintf("unkeyed:%d", index)
		}

		if _, exists := seen[key]; exists {
			removed++
			continue
		}

		seen[key] = struct{}{}
		result = append(result, item)
	}

	return result, removed
}

func uniqueAircraftTrajectories(
	items []trajectory.FlightTrajectory,
) ([]trajectory.FlightTrajectory, int) {
	result := make(
		[]trajectory.FlightTrajectory,
		0,
		len(items),
	)
	indexByKey := make(
		map[string]int,
		len(items),
	)
	removed := 0

	for index, item := range items {
		key, stable := aircraftContributorKey(item)
		if !stable {
			key = fmt.Sprintf(
				"unkeyed-aircraft:%d",
				index,
			)
		}

		existingIndex, exists := indexByKey[key]
		if !exists {
			indexByKey[key] = len(result)
			result = append(result, item)
			continue
		}

		removed++
		if trajectoryIsNewer(
			item,
			result[existingIndex],
		) {
			result[existingIndex] = item
		}
	}

	return result, removed
}

func aircraftContributorKey(
	item trajectory.FlightTrajectory,
) (string, bool) {
	if value := strings.TrimSpace(item.ICAO24); value != "" {
		return "icao24:" + strings.ToLower(value), true
	}

	if value := strings.TrimSpace(item.AircraftID); value != "" {
		return "aircraft-id:" + strings.ToLower(value), true
	}

	return trajectoryContributorKey(item)
}

func trajectoryIsNewer(
	candidate trajectory.FlightTrajectory,
	current trajectory.FlightTrajectory,
) bool {
	if candidate.EndTime.After(current.EndTime) {
		return true
	}

	return candidate.EndTime.Equal(current.EndTime) &&
		candidate.QualityScore > current.QualityScore
}

func trajectoryContributorKey(
	item trajectory.FlightTrajectory,
) (string, bool) {
	if value := strings.TrimSpace(item.IdentityKey); value != "" {
		return "identity:" + value, true
	}

	if value := strings.TrimSpace(item.FlightID); value != "" {
		return "flight:" + strings.ToLower(value), true
	}

	if value := strings.TrimSpace(item.ICAO24); value != "" {
		return "aircraft:" + strings.ToLower(value), true
	}

	return "", false
}
