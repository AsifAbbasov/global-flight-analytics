package projectionarrival

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"sort"
	"strings"
	"time"
)

func arrivalLimitations(
	mode EstimateMode,
	routeStatus routecontract.RouteStatus,
) []projectioncontract.Limitation {
	result := []projectioncontract.Limitation{
		{
			Code:    "arrival_radius_not_touchdown",
			Message: "Estimated arrival represents entry into the configured airport radius, not runway touchdown or gate arrival.",
			Scope:   "arrival",
		},
		{
			Code:    "destination_is_inferred",
			Message: "Destination airport is inferred by Route Intelligence and is not an official flight-plan destination.",
			Scope:   "arrival",
		},
		{
			Code:    "no_operational_arrival_intent",
			Message: "Official flight plan, Air Traffic Control sequence, runway assignment, holding, diversion, and pilot intent are unavailable.",
			Scope:   "arrival",
		},
		{
			Code:    "no_weather_arrival_adjustment",
			Message: "Weather and wind are not applied to the estimated arrival interval.",
			Scope:   "arrival",
		},
		{
			Code:    "research_only_arrival",
			Message: "Estimated arrival is a research output and must not be used for operational aviation decisions.",
			Scope:   "arrival",
		},
	}

	if mode == EstimateModeExtrapolated {
		result = append(
			result,
			projectioncontract.Limitation{
				Code:    "arrival_extrapolated_beyond_projection_horizon",
				Message: "Estimated arrival extends beyond the position-projection horizon using a bounded projected ground-speed profile.",
				Scope:   "arrival",
			},
		)
	}
	if routeStatus !=
		routecontract.RouteStatusComplete {
		result = append(
			result,
			projectioncontract.Limitation{
				Code:    "route_intelligence_partial",
				Message: "Route Intelligence resolved a destination without a complete two-endpoint route.",
				Scope:   "arrival",
			},
		)
	}

	return normalizeLimitations(result)
}

func normalizeLimitations(
	items []projectioncontract.Limitation,
) []projectioncontract.Limitation {
	seen := make(
		map[string]projectioncontract.Limitation,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message :=
			strings.TrimSpace(item.Message)
		scope := strings.TrimSpace(item.Scope)
		if code == "" ||
			message == "" ||
			scope == "" {
			continue
		}
		key := code + "\x00" +
			message + "\x00" +
			scope
		seen[key] =
			projectioncontract.Limitation{
				Code:    code,
				Message: message,
				Scope:   scope,
			}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]projectioncontract.Limitation,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func normalizeExplanations(
	items []projectioncontract.Explanation,
) []projectioncontract.Explanation {
	seen := make(
		map[string]projectioncontract.Explanation,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message :=
			strings.TrimSpace(item.Message)
		if code == "" ||
			message == "" {
			continue
		}
		key := code + "\x00" +
			message
		seen[key] =
			projectioncontract.Explanation{
				Code:    code,
				Message: message,
			}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]projectioncontract.Explanation,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func normalizeInputs(
	items []projectioncontract.InputReference,
) []projectioncontract.InputReference {
	type indexedInput struct {
		item  projectioncontract.InputReference
		index int
	}

	seen := make(
		map[string]indexedInput,
		len(items),
	)
	for index, item := range items {
		key :=
			strings.TrimSpace(item.Name) +
				"\x00" +
				string(item.Classification) +
				"\x00" +
				strings.TrimSpace(
					item.SourceName,
				) +
				"\x00" +
				item.ObservedAt.UTC().
					Format(time.RFC3339Nano)
		seen[key] = indexedInput{
			item:  item,
			index: index,
		}
	}

	values := make(
		[]indexedInput,
		0,
		len(seen),
	)
	for _, value := range seen {
		values = append(values, value)
	}
	sort.SliceStable(
		values,
		func(left int, right int) bool {
			return values[left].index <
				values[right].index
		},
	)

	result := make(
		[]projectioncontract.InputReference,
		0,
		len(values),
	)
	for _, value := range values {
		result = append(
			result,
			value.item,
		)
	}

	return result
}

func latestInputObservedAt(
	items []projectioncontract.InputReference,
) time.Time {
	var latest time.Time
	for _, item := range items {
		observedAt :=
			item.ObservedAt.UTC()
		if item.ObservedAt.IsZero() {
			continue
		}
		if latest.IsZero() ||
			observedAt.After(latest) {
			latest = observedAt
		}
	}

	return latest
}

func validateResult(
	result projectioncontract.Result,
) (projectioncontract.Result, error) {
	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrArrivalContractInvalid,
				report.Issues,
			)
	}

	return result.Clone(), nil
}
