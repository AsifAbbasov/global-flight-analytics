package trajectoryeligibility

import (
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const identityKeyPrefix = "flight-identity-"

type Evaluator struct {
	config Config
}

func New(
	config Config,
) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate trajectory analytics eligibility config: %w",
			err,
		)
	}

	return &Evaluator{
		config: config,
	}, nil
}

func NewDefault() *Evaluator {
	evaluator, err := New(
		DefaultConfig(),
	)
	if err != nil {
		panic(
			fmt.Sprintf(
				"default trajectory analytics eligibility config is invalid: %v",
				err,
			),
		)
	}

	return evaluator
}

func (
	evaluator *Evaluator,
) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) Evaluation {
	result := Evaluation{
		Decisions: make(
			[]Decision,
			0,
			len(orderedCapabilities),
		),
	}

	for _, capability := range orderedCapabilities {
		decision := evaluator.evaluateCapability(
			item,
			now,
			capability,
			evaluator.config.policy(
				capability,
			),
		)

		result.Decisions = append(
			result.Decisions,
			decision,
		)
		result.Permissions.set(
			capability,
			decision.Allowed,
		)
	}

	return result
}

func (
	evaluator *Evaluator,
) evaluateCapability(
	item trajectory.FlightTrajectory,
	now time.Time,
	capability Capability,
	policy Policy,
) Decision {
	reasons := make(
		[]ReasonCode,
		0,
		8,
	)

	if strings.TrimSpace(
		item.ICAO24,
	) == "" {
		reasons = appendReason(
			reasons,
			ReasonMissingAircraftIdentifier,
		)
	}

	duration, validTimeRange :=
		trajectoryDuration(item)

	if !validTimeRange {
		reasons = appendReason(
			reasons,
			ReasonInvalidTimeRange,
		)
	}

	if effectivePointCount(item) <
		policy.MinimumPointCount {
		reasons = appendReason(
			reasons,
			ReasonInsufficientPoints,
		)
	}

	if !isFiniteUnitScore(
		item.QualityScore,
	) ||
		item.QualityScore <
			policy.MinimumQualityScore {
		reasons = appendReason(
			reasons,
			ReasonLowQualityScore,
		)
	}

	if exceedsCoverageGapLimit(
		item,
		policy.MaximumCoverageGapCount,
	) {
		reasons = appendReason(
			reasons,
			ReasonTooManyCoverageGaps,
		)
	}

	if validTimeRange &&
		policy.MinimumDuration > 0 &&
		duration <
			policy.MinimumDuration {
		reasons = appendReason(
			reasons,
			ReasonDurationTooShort,
		)
	}

	if validTimeRange &&
		policy.MaximumDuration > 0 &&
		duration >
			policy.MaximumDuration {
		reasons = appendReason(
			reasons,
			ReasonDurationTooLong,
		)
	}

	if policy.RequireReliableIdentity {
		complete, reliable :=
			identityReliability(item)

		if !complete {
			reasons = appendReason(
				reasons,
				ReasonMissingIdentity,
			)
		} else if !reliable {
			reasons = appendReason(
				reasons,
				ReasonIdentityNotReliable,
			)
		}
	}

	if policy.RequireCallsign &&
		normalizeCallsign(
			item.Callsign,
		) == "" {
		reasons = appendReason(
			reasons,
			ReasonMissingCallsign,
		)
	}

	if policy.RequireAltitude &&
		!hasRecentUsableAltitude(
			item.Points,
		) {
		reasons = appendReason(
			reasons,
			ReasonMissingAltitude,
		)
	}

	if policy.MaximumObservationAge > 0 {
		if now.IsZero() {
			reasons = appendReason(
				reasons,
				ReasonEvaluationTimeMissing,
			)
		} else if !item.EndTime.IsZero() &&
			now.UTC().
				Sub(item.EndTime.UTC()) >
				policy.MaximumObservationAge {
			reasons = appendReason(
				reasons,
				ReasonStaleObservations,
			)
		}
	}

	if policy.MaximumRecentPointGap > 0 &&
		!hasRecentContinuity(
			item.Points,
			policy.MaximumRecentPointGap,
		) {
		reasons = appendReason(
			reasons,
			ReasonInsufficientRecentContinuity,
		)
	}

	return Decision{
		Capability: capability,
		Allowed:    len(reasons) == 0,
		Reasons:    reasons,
	}
}

func appendReason(
	reasons []ReasonCode,
	reason ReasonCode,
) []ReasonCode {
	for _, current := range reasons {
		if current == reason {
			return reasons
		}
	}

	return append(
		reasons,
		reason,
	)
}

func trajectoryDuration(
	item trajectory.FlightTrajectory,
) (time.Duration, bool) {
	if item.StartTime.IsZero() ||
		item.EndTime.IsZero() ||
		item.EndTime.Before(item.StartTime) {
		return 0, false
	}

	return item.EndTime.Sub(
		item.StartTime,
	), true
}

func effectivePointCount(
	item trajectory.FlightTrajectory,
) int {
	count := item.PointCount

	if len(item.Points) > count {
		count = len(item.Points)
	}

	return count
}

func effectiveCoverageGapCount(
	item trajectory.FlightTrajectory,
) int {
	count := item.CoverageGapCount

	if len(item.CoverageGaps) > count {
		count = len(item.CoverageGaps)
	}

	return count
}

func exceedsCoverageGapLimit(
	item trajectory.FlightTrajectory,
	maximum int,
) bool {
	count := effectiveCoverageGapCount(
		item,
	)

	if count < 0 {
		return true
	}

	return maximum >= 0 &&
		count > maximum
}

func isFiniteUnitScore(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}

func identityReliability(
	item trajectory.FlightTrajectory,
) (bool, bool) {
	if !isIdentityKey(
		item.IdentityKey,
	) ||
		!isKnownIdentityBasis(
			item.IdentityBasis,
		) ||
		!isKnownSplitReason(
			item.SplitReason,
		) {
		return false, false
	}

	switch item.IdentityBasis {
	case trajectory.FlightIdentityBasisSourceFlightID:
		return true,
			normalizeUUID(
				item.FlightID,
			) != ""

	case trajectory.FlightIdentityBasisCallsignAndStartTime:
		return true,
			normalizeCallsign(
				item.Callsign,
			) != ""

	case trajectory.FlightIdentityBasisAircraftAndStartTime:
		return true, false

	default:
		return false, false
	}
}

func isIdentityKey(
	value string,
) bool {
	if !strings.HasPrefix(
		value,
		identityKeyPrefix,
	) {
		return false
	}

	digest := strings.TrimPrefix(
		value,
		identityKeyPrefix,
	)

	if len(digest) != 64 ||
		digest != strings.ToLower(digest) {
		return false
	}

	decoded, err := hex.DecodeString(
		digest,
	)

	return err == nil &&
		len(decoded) == 32
}

func isKnownIdentityBasis(
	value trajectory.FlightIdentityBasis,
) bool {
	switch value {
	case trajectory.FlightIdentityBasisSourceFlightID,
		trajectory.FlightIdentityBasisCallsignAndStartTime,
		trajectory.FlightIdentityBasisAircraftAndStartTime:
		return true

	default:
		return false
	}
}

func isKnownSplitReason(
	value trajectory.FlightSplitReason,
) bool {
	switch value {
	case trajectory.FlightSplitReasonInitialObservation,
		trajectory.FlightSplitReasonSourceFlightIDChanged,
		trajectory.FlightSplitReasonCallsignChanged,
		trajectory.FlightSplitReasonGroundCycle,
		trajectory.FlightSplitReasonContinuedFromPreviousBatch:
		return true

	default:
		return false
	}
}

func normalizeCallsign(
	value string,
) string {
	return strings.ToUpper(
		strings.TrimSpace(value),
	)
}

func normalizeUUID(
	value string,
) string {
	normalized := strings.ToLower(
		strings.TrimSpace(value),
	)

	if len(normalized) != 36 {
		return ""
	}

	for index, character := range normalized {
		switch index {
		case 8, 13, 18, 23:
			if character != '-' {
				return ""
			}

		default:
			if character < '0' ||
				character > '9' &&
					(character < 'a' ||
						character > 'f') {
				return ""
			}
		}
	}

	return normalized
}

func hasRecentUsableAltitude(
	points []trajectory.TrackPoint4D,
) bool {
	recent := latestPoints(
		points,
		2,
	)

	if len(recent) < 2 {
		return false
	}

	for _, point := range recent {
		if !hasUsableAltitude(
			point,
		) {
			return false
		}
	}

	return true
}

func hasUsableAltitude(
	point trajectory.TrackPoint4D,
) bool {
	barometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
		)
	geometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
		)

	return isUsableAltitudeStatus(
		barometricStatus,
	) ||
		isUsableAltitudeStatus(
			geometricStatus,
		)
}

func isUsableAltitudeStatus(
	status flightstate.AltitudeStatus,
) bool {
	return status ==
		flightstate.AltitudeStatusObserved ||
		status ==
			flightstate.AltitudeStatusGround
}

func hasRecentContinuity(
	points []trajectory.TrackPoint4D,
	maximumGap time.Duration,
) bool {
	recent := latestPoints(
		points,
		2,
	)

	if len(recent) < 2 {
		return false
	}

	gap := recent[1].ObservedAt.Sub(
		recent[0].ObservedAt,
	)

	return gap > 0 &&
		gap <= maximumGap
}

func latestPoints(
	points []trajectory.TrackPoint4D,
	count int,
) []trajectory.TrackPoint4D {
	if count <= 0 ||
		len(points) == 0 {
		return nil
	}

	sorted := append(
		[]trajectory.TrackPoint4D(nil),
		points...,
	)

	sort.SliceStable(
		sorted,
		func(
			left int,
			right int,
		) bool {
			if sorted[left].ObservedAt.Equal(
				sorted[right].ObservedAt,
			) {
				return sorted[left].ID <
					sorted[right].ID
			}

			return sorted[left].ObservedAt.Before(
				sorted[right].ObservedAt,
			)
		},
	)

	if len(sorted) <= count {
		return sorted
	}

	return append(
		[]trajectory.TrackPoint4D(nil),
		sorted[len(sorted)-count:]...,
	)
}
