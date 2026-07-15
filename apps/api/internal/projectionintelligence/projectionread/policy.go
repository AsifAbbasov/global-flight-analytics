package projectionread

import (
	"fmt"
	"time"
)

const (
	Version = "projection-production-read-service-v1"

	DefaultSourceName          = "postgres_projection_intelligence_read_v1"
	DefaultSimilarityPolicyKey = "route_scoped_local_continuation_v1"
)

type DataSourcePolicy struct {
	MaximumTrajectoryPointCount     int
	MaximumHistoricalCandidateCount int
	HistoricalCandidateLookback     time.Duration
	RouteHistoryWindow              time.Duration
	RecentRouteWindow               time.Duration
	SourceName                      string
}

type HorizonPolicy struct {
	Name              string
	MinimumDuration   time.Duration
	DefaultDuration   time.Duration
	MaximumDuration   time.Duration
	Step              time.Duration
	MaximumPointCount int
}

type KinematicPolicy struct {
	InitialHorizontalUncertaintyM  float64
	HorizontalUncertaintyGrowthMPS float64
	InitialVerticalUncertaintyM    float64
	VerticalUncertaintyGrowthMPS   float64
	MaximumConfidenceLoss          float64
	MediumConfidenceMinimum        float64
	HighConfidenceMinimum          float64
	AllowOnGround                  bool
}

type NeighborPolicy struct {
	SimilarityPolicyKey      string
	MinimumCurrentPointCount int
	MaximumCandidateCount    int
	SelectionLimit           int
	MinimumSimilarityScore   float64
	MaximumAnchorDistanceKM  float64
	MaximumCandidateAge      time.Duration
}

type PatternPolicy struct {
	MinimumNeighborCount        int
	TargetNeighborCount         int
	MaximumCandidateAge         time.Duration
	MaximumMeanAnchorDistanceKM float64
	MinimumUsableScore          float64
	MediumConfidenceMinimum     float64
	HighConfidenceMinimum       float64
	SimilarityWeight            float64
	SupportWeight               float64
	FreshnessWeight             float64
	AnchorProximityWeight       float64
}

type FreshnessPolicy struct {
	MaximumNewestNeighborAge   time.Duration
	MaximumMeanNeighborAge     time.Duration
	MaximumOldestNeighborAge   time.Duration
	RecentNeighborAgeLimit     time.Duration
	MinimumRecentNeighborCount int
	TargetRecentNeighborCount  int
	MinimumUsableScore         float64
	CompleteScoreMinimum       float64
	NewestAgeWeight            float64
	MeanAgeWeight              float64
	OldestAgeWeight            float64
	RecentSupportWeight        float64
}

type RouteFrequencyPolicy struct {
	MinimumObservationCount       int
	TargetObservationCount        int
	MinimumDistinctDayCount       int
	TargetDistinctDayCount        int
	RecentWindow                  time.Duration
	MinimumRecentObservationCount int
	TargetRecentObservationCount  int
	MaximumLatestObservationAge   time.Duration
	MinimumRouteConfidenceScore   float64
	MinimumUsableScore            float64
	CompleteScoreMinimum          float64
	ObservationCountWeight        float64
	DistinctDayWeight             float64
	RecentObservationWeight       float64
	LatestObservationWeight       float64
	RouteConfidenceWeight         float64
}

type ContinuationPolicy struct {
	MinimumPointSupport            int
	MinimumAltitudeSupport         int
	InitialHorizontalUncertaintyM  float64
	HorizontalUncertaintyGrowthMPS float64
	InitialVerticalUncertaintyM    float64
	VerticalUncertaintyGrowthMPS   float64
	NeighborSpreadMultiplier       float64
	MaximumConfidenceLoss          float64
	MediumConfidenceMinimum        float64
	HighConfidenceMinimum          float64
}

type ArrivalPolicy struct {
	ArrivalRadiusM                     float64
	MinimumDestinationConfidenceScore  float64
	MinimumSpeedSampleCount            int
	MaximumSpeedSampleCount            int
	MinimumGroundSpeedMPS              float64
	SpeedUncertaintyMultiplier         float64
	MinimumArrivalInterval             time.Duration
	MaximumEstimatedArrivalDuration    time.Duration
	MaximumExtrapolationConfidenceLoss float64
	ProjectionConfidenceWeight         float64
	DestinationConfidenceWeight        float64
	SpeedStabilityWeight               float64
	MediumConfidenceMinimum            float64
	HighConfidenceMinimum              float64
}

type Policy struct {
	DataSource     DataSourcePolicy
	Horizon        HorizonPolicy
	Kinematic      KinematicPolicy
	Neighbors      NeighborPolicy
	Pattern        PatternPolicy
	Freshness      FreshnessPolicy
	RouteFrequency RouteFrequencyPolicy
	Continuation   ContinuationPolicy
	Arrival        ArrivalPolicy
}

func DefaultPolicy() Policy {
	return Policy{
		DataSource: DataSourcePolicy{
			MaximumTrajectoryPointCount:     10000,
			MaximumHistoricalCandidateCount: 50,
			HistoricalCandidateLookback:     90 * 24 * time.Hour,
			RouteHistoryWindow:              180 * 24 * time.Hour,
			RecentRouteWindow:               30 * 24 * time.Hour,
			SourceName:                      DefaultSourceName,
		},
		Horizon: HorizonPolicy{
			Name:              "stage-9-production-short-horizon-v1",
			MinimumDuration:   time.Minute,
			DefaultDuration:   5 * time.Minute,
			MaximumDuration:   15 * time.Minute,
			Step:              30 * time.Second,
			MaximumPointCount: 30,
		},
		Kinematic: KinematicPolicy{
			InitialHorizontalUncertaintyM:  500,
			HorizontalUncertaintyGrowthMPS: 8,
			InitialVerticalUncertaintyM:    150,
			VerticalUncertaintyGrowthMPS:   1,
			MaximumConfidenceLoss:          0.45,
			MediumConfidenceMinimum:        0.55,
			HighConfidenceMinimum:          0.80,
			AllowOnGround:                  false,
		},
		Neighbors: NeighborPolicy{
			SimilarityPolicyKey:      DefaultSimilarityPolicyKey,
			MinimumCurrentPointCount: 5,
			MaximumCandidateCount:    50,
			SelectionLimit:           5,
			MinimumSimilarityScore:   0.60,
			MaximumAnchorDistanceKM:  100,
			MaximumCandidateAge:      90 * 24 * time.Hour,
		},
		Pattern: PatternPolicy{
			MinimumNeighborCount:        2,
			TargetNeighborCount:         5,
			MaximumCandidateAge:         90 * 24 * time.Hour,
			MaximumMeanAnchorDistanceKM: 100,
			MinimumUsableScore:          0.55,
			MediumConfidenceMinimum:     0.60,
			HighConfidenceMinimum:       0.80,
			SimilarityWeight:            0.45,
			SupportWeight:               0.20,
			FreshnessWeight:             0.20,
			AnchorProximityWeight:       0.15,
		},
		Freshness: FreshnessPolicy{
			MaximumNewestNeighborAge:   30 * 24 * time.Hour,
			MaximumMeanNeighborAge:     60 * 24 * time.Hour,
			MaximumOldestNeighborAge:   90 * 24 * time.Hour,
			RecentNeighborAgeLimit:     30 * 24 * time.Hour,
			MinimumRecentNeighborCount: 1,
			TargetRecentNeighborCount:  3,
			MinimumUsableScore:         0.45,
			CompleteScoreMinimum:       0.70,
			NewestAgeWeight:            0.30,
			MeanAgeWeight:              0.30,
			OldestAgeWeight:            0.20,
			RecentSupportWeight:        0.20,
		},
		RouteFrequency: RouteFrequencyPolicy{
			MinimumObservationCount:       3,
			TargetObservationCount:        10,
			MinimumDistinctDayCount:       2,
			TargetDistinctDayCount:        7,
			RecentWindow:                  30 * 24 * time.Hour,
			MinimumRecentObservationCount: 1,
			TargetRecentObservationCount:  4,
			MaximumLatestObservationAge:   30 * 24 * time.Hour,
			MinimumRouteConfidenceScore:   0.60,
			MinimumUsableScore:            0.45,
			CompleteScoreMinimum:          0.75,
			ObservationCountWeight:        0.25,
			DistinctDayWeight:             0.20,
			RecentObservationWeight:       0.20,
			LatestObservationWeight:       0.20,
			RouteConfidenceWeight:         0.15,
		},
		Continuation: ContinuationPolicy{
			MinimumPointSupport:            2,
			MinimumAltitudeSupport:         1,
			InitialHorizontalUncertaintyM:  750,
			HorizontalUncertaintyGrowthMPS: 10,
			InitialVerticalUncertaintyM:    200,
			VerticalUncertaintyGrowthMPS:   1.5,
			NeighborSpreadMultiplier:       1.5,
			MaximumConfidenceLoss:          0.50,
			MediumConfidenceMinimum:        0.55,
			HighConfidenceMinimum:          0.80,
		},
		Arrival: ArrivalPolicy{
			ArrivalRadiusM:                     10000,
			MinimumDestinationConfidenceScore:  0.60,
			MinimumSpeedSampleCount:            3,
			MaximumSpeedSampleCount:            8,
			MinimumGroundSpeedMPS:              30,
			SpeedUncertaintyMultiplier:         1.5,
			MinimumArrivalInterval:             2 * time.Minute,
			MaximumEstimatedArrivalDuration:    8 * time.Hour,
			MaximumExtrapolationConfidenceLoss: 0.50,
			ProjectionConfidenceWeight:         0.45,
			DestinationConfidenceWeight:        0.35,
			SpeedStabilityWeight:               0.20,
			MediumConfidenceMinimum:            0.55,
			HighConfidenceMinimum:              0.80,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.DataSource.MaximumTrajectoryPointCount < 2 {
		return fmt.Errorf(
			"maximum trajectory point count must be at least two",
		)
	}
	if policy.DataSource.MaximumHistoricalCandidateCount < 1 {
		return fmt.Errorf(
			"maximum historical candidate count must be positive",
		)
	}
	if policy.DataSource.HistoricalCandidateLookback <= 0 ||
		policy.DataSource.RouteHistoryWindow <= 0 ||
		policy.DataSource.RecentRouteWindow <= 0 ||
		policy.DataSource.RecentRouteWindow >
			policy.DataSource.RouteHistoryWindow {
		return fmt.Errorf(
			"projection read data-source windows are invalid",
		)
	}
	if policy.DataSource.SourceName == "" {
		return fmt.Errorf(
			"projection read source name is required",
		)
	}
	if policy.Neighbors.MaximumCandidateCount <
		policy.DataSource.MaximumHistoricalCandidateCount {
		return fmt.Errorf(
			"neighbor maximum candidate count must cover the data-source candidate limit",
		)
	}
	if policy.Neighbors.MaximumCandidateAge !=
		policy.DataSource.HistoricalCandidateLookback {
		return fmt.Errorf(
			"candidate lookback and neighbor maximum age must match",
		)
	}
	if policy.RouteFrequency.RecentWindow !=
		policy.DataSource.RecentRouteWindow {
		return fmt.Errorf(
			"route-frequency and data-source recent windows must match",
		)
	}
	return nil
}
