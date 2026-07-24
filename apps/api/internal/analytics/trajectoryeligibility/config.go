package trajectoryeligibility

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const DefaultMaximumFutureObservationSkew = 30 * time.Second

var (
	ErrMinimumPointCountInvalid = errors.New(
		"minimum point count must be non-negative",
	)
	ErrMinimumQualityScoreInvalid = errors.New(
		"minimum quality score must be finite and between zero and one",
	)
	ErrMaximumCoverageGapCountInvalid = errors.New(
		"maximum coverage gap count must be minus one or greater",
	)
	ErrMinimumDurationInvalid = errors.New(
		"minimum duration must be non-negative",
	)
	ErrMaximumDurationInvalid = errors.New(
		"maximum duration must be non-negative",
	)
	ErrDurationRangeInvalid = errors.New(
		"maximum duration must be zero or greater than or equal to minimum duration",
	)
	ErrMaximumObservationAgeInvalid = errors.New(
		"maximum observation age must be non-negative",
	)
	ErrMaximumFutureObservationSkewInvalid = errors.New(
		"maximum future observation skew must be non-negative",
	)
	ErrMaximumRecentPointGapInvalid = errors.New(
		"maximum recent point gap must be non-negative",
	)
)

type Policy struct {
	MinimumPointCount            int
	MinimumQualityScore          float64
	MaximumCoverageGapCount      int
	MinimumDuration              time.Duration
	MaximumDuration              time.Duration
	MaximumObservationAge        time.Duration
	MaximumFutureObservationSkew time.Duration
	MaximumRecentPointGap        time.Duration
	RequireReliableIdentity      bool
	RequireCallsign              bool
	RequireAltitude              bool
}

func (
	policy Policy,
) Validate() error {
	if policy.MinimumPointCount < 0 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumPointCountInvalid,
			policy.MinimumPointCount,
		)
	}

	if math.IsNaN(
		policy.MinimumQualityScore,
	) ||
		math.IsInf(
			policy.MinimumQualityScore,
			0,
		) ||
		policy.MinimumQualityScore < 0 ||
		policy.MinimumQualityScore > 1 {
		return fmt.Errorf(
			"%w: %f",
			ErrMinimumQualityScoreInvalid,
			policy.MinimumQualityScore,
		)
	}

	if policy.MaximumCoverageGapCount < -1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMaximumCoverageGapCountInvalid,
			policy.MaximumCoverageGapCount,
		)
	}

	if policy.MinimumDuration < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMinimumDurationInvalid,
			policy.MinimumDuration,
		)
	}

	if policy.MaximumDuration < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumDurationInvalid,
			policy.MaximumDuration,
		)
	}

	if policy.MaximumDuration > 0 &&
		policy.MaximumDuration <
			policy.MinimumDuration {
		return fmt.Errorf(
			"%w: minimum=%s maximum=%s",
			ErrDurationRangeInvalid,
			policy.MinimumDuration,
			policy.MaximumDuration,
		)
	}

	if policy.MaximumObservationAge < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumObservationAgeInvalid,
			policy.MaximumObservationAge,
		)
	}

	if policy.MaximumFutureObservationSkew < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumFutureObservationSkewInvalid,
			policy.MaximumFutureObservationSkew,
		)
	}

	if policy.MaximumRecentPointGap < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumRecentPointGapInvalid,
			policy.MaximumRecentPointGap,
		)
	}

	return nil
}

type Config struct {
	TrafficMetrics        Policy
	AirportActivity       Policy
	RouteInference        Policy
	HistoricalAggregation Policy
	Projection            Policy
}

func DefaultConfig() Config {
	return Config{
		TrafficMetrics: Policy{
			MinimumPointCount:            1,
			MinimumQualityScore:          0.20,
			MaximumCoverageGapCount:      5,
			MaximumObservationAge:        5 * time.Minute,
			MaximumFutureObservationSkew: DefaultMaximumFutureObservationSkew,
		},
		AirportActivity: Policy{
			MinimumPointCount:            2,
			MinimumQualityScore:          0.40,
			MaximumCoverageGapCount:      2,
			MaximumObservationAge:        15 * time.Minute,
			MaximumFutureObservationSkew: DefaultMaximumFutureObservationSkew,
			RequireReliableIdentity:      true,
		},
		RouteInference: Policy{
			MinimumPointCount:            4,
			MinimumQualityScore:          0.60,
			MaximumCoverageGapCount:      1,
			MinimumDuration:              2 * time.Minute,
			MaximumDuration:              18 * time.Hour,
			MaximumFutureObservationSkew: DefaultMaximumFutureObservationSkew,
			RequireReliableIdentity:      true,
		},
		HistoricalAggregation: Policy{
			MinimumPointCount:            3,
			MinimumQualityScore:          0.50,
			MaximumCoverageGapCount:      3,
			MinimumDuration:              time.Minute,
			MaximumDuration:              24 * time.Hour,
			MaximumFutureObservationSkew: DefaultMaximumFutureObservationSkew,
			RequireReliableIdentity:      true,
		},
		Projection: Policy{
			MinimumPointCount:            5,
			MinimumQualityScore:          0.75,
			MaximumCoverageGapCount:      0,
			MinimumDuration:              2 * time.Minute,
			MaximumDuration:              18 * time.Hour,
			MaximumObservationAge:        2 * time.Minute,
			MaximumFutureObservationSkew: DefaultMaximumFutureObservationSkew,
			MaximumRecentPointGap:        90 * time.Second,
			RequireReliableIdentity:      true,
			RequireAltitude:              true,
		},
	}
}

func (
	config Config,
) Validate() error {
	for _, item := range []struct {
		name   Capability
		policy Policy
	}{
		{
			name:   CapabilityTrafficMetrics,
			policy: config.TrafficMetrics,
		},
		{
			name:   CapabilityAirportActivity,
			policy: config.AirportActivity,
		},
		{
			name:   CapabilityRouteInference,
			policy: config.RouteInference,
		},
		{
			name:   CapabilityHistoricalAggregation,
			policy: config.HistoricalAggregation,
		},
		{
			name:   CapabilityProjection,
			policy: config.Projection,
		},
	} {
		if err := item.policy.Validate(); err != nil {
			return fmt.Errorf(
				"validate %s eligibility policy: %w",
				item.name,
				err,
			)
		}
	}

	return nil
}

func (
	config Config,
) policy(
	capability Capability,
) Policy {
	switch capability {
	case CapabilityTrafficMetrics:
		return config.TrafficMetrics

	case CapabilityAirportActivity:
		return config.AirportActivity

	case CapabilityRouteInference:
		return config.RouteInference

	case CapabilityHistoricalAggregation:
		return config.HistoricalAggregation

	case CapabilityProjection:
		return config.Projection

	default:
		return Policy{}
	}
}
