package projectionproduction

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
)

const (
	Version            = "projection-production-composition-v1"
	FingerprintVersion = "projection-production-composition-fingerprint-v1"
)

type Strategy string

const (
	StrategyKinematic          Strategy = "kinematic_baseline"
	StrategyHistoricalNeighbor Strategy = "historical_neighbor_continuation"
)

func (strategy Strategy) IsKnown() bool {
	switch strategy {
	case StrategyKinematic,
		StrategyHistoricalNeighbor:
		return true
	default:
		return false
	}
}

type ArrivalStatus string

const (
	ArrivalStatusAttached ArrivalStatus = "attached"
	ArrivalStatusWithheld ArrivalStatus = "withheld"
	ArrivalStatusFailed   ArrivalStatus = "failed"
	ArrivalStatusSkipped  ArrivalStatus = "skipped"
)

func (status ArrivalStatus) IsKnown() bool {
	switch status {
	case ArrivalStatusAttached,
		ArrivalStatusWithheld,
		ArrivalStatusFailed,
		ArrivalStatusSkipped:
		return true
	default:
		return false
	}
}

type Notice struct {
	Code    string
	Message string
}

type Result struct {
	Version string

	Strategy       Strategy
	FallbackReason string
	ArrivalStatus  ArrivalStatus

	Projection projectioncontract.Result

	NeighborSelection *projectionneighbors.Result
	PatternConfidence *projectionpatternconfidence.Result
	Freshness         *projectionfreshness.Result
	RouteFrequency    *projectionroutefrequency.Result

	Notices []Notice

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Projection = result.Projection.Clone()
	if result.NeighborSelection != nil {
		value := result.NeighborSelection.Clone()
		cloned.NeighborSelection = &value
	}
	if result.PatternConfidence != nil {
		value := result.PatternConfidence.Clone()
		cloned.PatternConfidence = &value
	}
	if result.Freshness != nil {
		value := result.Freshness.Clone()
		cloned.Freshness = &value
	}
	if result.RouteFrequency != nil {
		value := result.RouteFrequency.Clone()
		cloned.RouteFrequency = &value
	}
	cloned.Notices = append(
		[]Notice(nil),
		result.Notices...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version ||
		!result.Strategy.IsKnown() ||
		!result.ArrivalStatus.IsKnown() {
		return fmt.Errorf(
			"production composition metadata is invalid",
		)
	}
	if result.GeneratedAt.IsZero() ||
		!result.GeneratedAt.Equal(
			result.Projection.GeneratedAt,
		) {
		return fmt.Errorf(
			"production composition generated-at time is invalid",
		)
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"production composition fingerprint is invalid",
		)
	}

	projectionReport := projectioncontract.Validate(
		result.Projection,
	)
	if projectionReport.Status !=
		projectioncontract.ValidationStatusValid {
		return fmt.Errorf(
			"production projection contract is invalid: %#v",
			projectionReport.Issues,
		)
	}

	if result.NeighborSelection != nil {
		if err := result.NeighborSelection.Validate(); err != nil {
			return fmt.Errorf(
				"production neighbor selection is invalid: %w",
				err,
			)
		}
	}
	if result.PatternConfidence != nil {
		if err := result.PatternConfidence.Validate(); err != nil {
			return fmt.Errorf(
				"production pattern confidence is invalid: %w",
				err,
			)
		}
	}
	if result.Freshness != nil {
		if err := result.Freshness.Validate(); err != nil {
			return fmt.Errorf(
				"production freshness result is invalid: %w",
				err,
			)
		}
	}
	if result.RouteFrequency != nil {
		if err := result.RouteFrequency.Validate(); err != nil {
			return fmt.Errorf(
				"production route-frequency result is invalid: %w",
				err,
			)
		}
	}

	switch result.Strategy {
	case StrategyHistoricalNeighbor:
		if strings.TrimSpace(result.FallbackReason) != "" ||
			result.Projection.Method.Name !=
				projectioncontinuation.MethodName ||
			result.NeighborSelection == nil ||
			result.PatternConfidence == nil ||
			result.Freshness == nil ||
			result.RouteFrequency == nil ||
			!result.PatternConfidence.Usable ||
			!result.Freshness.Usable ||
			!result.RouteFrequency.Usable {
			return fmt.Errorf(
				"historical production strategy does not contain complete usable evidence",
			)
		}
	case StrategyKinematic:
		if strings.TrimSpace(result.FallbackReason) == "" ||
			result.Projection.Method.Name !=
				projectionbaseline.MethodName {
			return fmt.Errorf(
				"kinematic production strategy requires a fallback reason and kinematic method",
			)
		}
	}

	switch result.ArrivalStatus {
	case ArrivalStatusAttached:
		if result.Projection.Arrival == nil {
			return fmt.Errorf(
				"attached arrival status requires an arrival estimate",
			)
		}
	case ArrivalStatusWithheld,
		ArrivalStatusFailed,
		ArrivalStatusSkipped:
		if result.Projection.Arrival != nil {
			return fmt.Errorf(
				"non-attached arrival status must not contain an arrival estimate",
			)
		}
	}

	for _, notice := range result.Notices {
		if strings.TrimSpace(notice.Code) == "" ||
			strings.TrimSpace(notice.Message) == "" {
			return fmt.Errorf(
				"production composition notice is invalid",
			)
		}
	}
	if result.Strategy == StrategyKinematic &&
		len(result.Notices) == 0 {
		return fmt.Errorf(
			"kinematic production fallback requires an auditable notice",
		)
	}

	return nil
}

func normalizeNotices(
	items []Notice,
) []Notice {
	seen := make(
		map[string]Notice,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		seen[key] = Notice{
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
		[]Notice,
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
