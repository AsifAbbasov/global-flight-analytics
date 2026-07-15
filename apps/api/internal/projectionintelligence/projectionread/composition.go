package projectionread

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Pool   *pgxpool.Pool
	Policy Policy
	Now    func() time.Time
}

type PostgresComposition struct {
	Service    *Service
	DataSource *PostgresDataSource
	Composer   *projectionproduction.Composer
	Policy     Policy
}

func NewPostgres(
	config PostgresConfig,
) (*PostgresComposition, error) {
	if config.Pool == nil {
		return nil,
			fmt.Errorf(
				"Projection Intelligence PostgreSQL pool is required",
			)
	}
	if err := config.Policy.Validate(); err != nil {
		return nil,
			fmt.Errorf(
				"validate Projection Intelligence production policy: %w",
				err,
			)
	}

	components, err :=
		buildAlgorithmComponents(
			config.Policy,
		)
	if err != nil {
		return nil,
			err
	}

	dataSource, err :=
		NewPostgresDataSource(
			PostgresDataSourceConfig{
				Pool:   config.Pool,
				Policy: config.Policy.DataSource,
			},
		)
	if err != nil {
		return nil,
			fmt.Errorf(
				"construct Projection Intelligence PostgreSQL data source: %w",
				err,
			)
	}

	service, err := NewService(
		ServiceConfig{
			DataSource: dataSource,
			Composer:   components.composer,
			Policy:     config.Policy,
			Now:        config.Now,
		},
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"construct Projection Intelligence read service: %w",
				err,
			)
	}

	return &PostgresComposition{
		Service:    service,
		DataSource: dataSource,
		Composer:   components.composer,
		Policy:     config.Policy,
	}, nil
}

type algorithmComponents struct {
	composer *projectionproduction.Composer
}

func buildAlgorithmComponents(
	policy Policy,
) (algorithmComponents, error) {
	horizonPlanner, err :=
		projectionhorizon.New(
			projectionhorizon.Config{
				Name: policy.Horizon.Name,
				MinimumDuration: policy.Horizon.
					MinimumDuration,
				DefaultDuration: policy.Horizon.
					DefaultDuration,
				MaximumDuration: policy.Horizon.
					MaximumDuration,
				Step: policy.Horizon.Step,
				MaximumPointCount: policy.Horizon.
					MaximumPointCount,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"horizon planner",
				err,
			)
	}

	eligibility :=
		trajectoryeligibility.NewDefault()

	kinematic, err :=
		projectionbaseline.New(
			projectionbaseline.Config{
				HorizonPlanner:       horizonPlanner,
				EligibilityEvaluator: eligibility,
				InitialHorizontalUncertaintyM: policy.Kinematic.
					InitialHorizontalUncertaintyM,
				HorizontalUncertaintyGrowthMPS: policy.Kinematic.
					HorizontalUncertaintyGrowthMPS,
				InitialVerticalUncertaintyM: policy.Kinematic.
					InitialVerticalUncertaintyM,
				VerticalUncertaintyGrowthMPS: policy.Kinematic.
					VerticalUncertaintyGrowthMPS,
				MaximumConfidenceLoss: policy.Kinematic.
					MaximumConfidenceLoss,
				MediumConfidenceMinimum: policy.Kinematic.
					MediumConfidenceMinimum,
				HighConfidenceMinimum: policy.Kinematic.
					HighConfidenceMinimum,
				AllowOnGround: policy.Kinematic.
					AllowOnGround,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"kinematic baseline",
				err,
			)
	}

	similarity :=
		historicalsimilarity.NewDefault()

	selector, err :=
		projectionneighbors.New(
			projectionneighbors.Config{
				SimilarityEngine: similarity,
				SimilarityPolicyKey: policy.Neighbors.
					SimilarityPolicyKey,
				MinimumCurrentPointCount: policy.Neighbors.
					MinimumCurrentPointCount,
				MaximumCandidateCount: policy.Neighbors.
					MaximumCandidateCount,
				SelectionLimit: policy.Neighbors.
					SelectionLimit,
				MinimumSimilarityScore: policy.Neighbors.
					MinimumSimilarityScore,
				MaximumAnchorDistanceKM: policy.Neighbors.
					MaximumAnchorDistanceKM,
				MaximumCandidateAge: policy.Neighbors.
					MaximumCandidateAge,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"historical neighbor selector",
				err,
			)
	}

	pattern, err :=
		projectionpatternconfidence.New(
			projectionpatternconfidence.Config{
				MinimumNeighborCount: policy.Pattern.
					MinimumNeighborCount,
				TargetNeighborCount: policy.Pattern.
					TargetNeighborCount,
				MaximumCandidateAge: policy.Pattern.
					MaximumCandidateAge,
				MaximumMeanAnchorDistanceKM: policy.Pattern.
					MaximumMeanAnchorDistanceKM,
				MinimumUsableScore: policy.Pattern.
					MinimumUsableScore,
				MediumConfidenceMinimum: policy.Pattern.
					MediumConfidenceMinimum,
				HighConfidenceMinimum: policy.Pattern.
					HighConfidenceMinimum,
				SimilarityWeight: policy.Pattern.
					SimilarityWeight,
				SupportWeight: policy.Pattern.
					SupportWeight,
				FreshnessWeight: policy.Pattern.
					FreshnessWeight,
				AnchorProximityWeight: policy.Pattern.
					AnchorProximityWeight,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"pattern confidence evaluator",
				err,
			)
	}

	freshness, err :=
		projectionfreshness.New(
			projectionfreshness.Config{
				MaximumNewestNeighborAge: policy.Freshness.
					MaximumNewestNeighborAge,
				MaximumMeanNeighborAge: policy.Freshness.
					MaximumMeanNeighborAge,
				MaximumOldestNeighborAge: policy.Freshness.
					MaximumOldestNeighborAge,
				RecentNeighborAgeLimit: policy.Freshness.
					RecentNeighborAgeLimit,
				MinimumRecentNeighborCount: policy.Freshness.
					MinimumRecentNeighborCount,
				TargetRecentNeighborCount: policy.Freshness.
					TargetRecentNeighborCount,
				MinimumUsableScore: policy.Freshness.
					MinimumUsableScore,
				CompleteScoreMinimum: policy.Freshness.
					CompleteScoreMinimum,
				NewestAgeWeight: policy.Freshness.
					NewestAgeWeight,
				MeanAgeWeight: policy.Freshness.
					MeanAgeWeight,
				OldestAgeWeight: policy.Freshness.
					OldestAgeWeight,
				RecentSupportWeight: policy.Freshness.
					RecentSupportWeight,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"pattern freshness evaluator",
				err,
			)
	}

	frequency, err :=
		projectionroutefrequency.New(
			projectionroutefrequency.Config{
				MinimumObservationCount: policy.RouteFrequency.
					MinimumObservationCount,
				TargetObservationCount: policy.RouteFrequency.
					TargetObservationCount,
				MinimumDistinctDayCount: policy.RouteFrequency.
					MinimumDistinctDayCount,
				TargetDistinctDayCount: policy.RouteFrequency.
					TargetDistinctDayCount,
				RecentWindow: policy.RouteFrequency.
					RecentWindow,
				MinimumRecentObservationCount: policy.RouteFrequency.
					MinimumRecentObservationCount,
				TargetRecentObservationCount: policy.RouteFrequency.
					TargetRecentObservationCount,
				MaximumLatestObservationAge: policy.RouteFrequency.
					MaximumLatestObservationAge,
				MinimumRouteConfidenceScore: policy.RouteFrequency.
					MinimumRouteConfidenceScore,
				MinimumUsableScore: policy.RouteFrequency.
					MinimumUsableScore,
				CompleteScoreMinimum: policy.RouteFrequency.
					CompleteScoreMinimum,
				ObservationCountWeight: policy.RouteFrequency.
					ObservationCountWeight,
				DistinctDayWeight: policy.RouteFrequency.
					DistinctDayWeight,
				RecentObservationWeight: policy.RouteFrequency.
					RecentObservationWeight,
				LatestObservationWeight: policy.RouteFrequency.
					LatestObservationWeight,
				RouteConfidenceWeight: policy.RouteFrequency.
					RouteConfidenceWeight,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"route-frequency evaluator",
				err,
			)
	}

	historical, err :=
		projectioncontinuation.New(
			projectioncontinuation.Config{
				HorizonPlanner:             horizonPlanner,
				NeighborSelector:           selector,
				PatternConfidenceEvaluator: pattern,
				FallbackProjector:          kinematic,
				MinimumPointSupport: policy.Continuation.
					MinimumPointSupport,
				MinimumAltitudeSupport: policy.Continuation.
					MinimumAltitudeSupport,
				InitialHorizontalUncertaintyM: policy.Continuation.
					InitialHorizontalUncertaintyM,
				HorizontalUncertaintyGrowthMPS: policy.Continuation.
					HorizontalUncertaintyGrowthMPS,
				InitialVerticalUncertaintyM: policy.Continuation.
					InitialVerticalUncertaintyM,
				VerticalUncertaintyGrowthMPS: policy.Continuation.
					VerticalUncertaintyGrowthMPS,
				NeighborSpreadMultiplier: policy.Continuation.
					NeighborSpreadMultiplier,
				MaximumConfidenceLoss: policy.Continuation.
					MaximumConfidenceLoss,
				MediumConfidenceMinimum: policy.Continuation.
					MediumConfidenceMinimum,
				HighConfidenceMinimum: policy.Continuation.
					HighConfidenceMinimum,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"historical continuation projector",
				err,
			)
	}

	arrival, err :=
		projectionarrival.New(
			projectionarrival.Config{
				ArrivalRadiusM: policy.Arrival.
					ArrivalRadiusM,
				MinimumDestinationConfidenceScore: policy.Arrival.
					MinimumDestinationConfidenceScore,
				MinimumSpeedSampleCount: policy.Arrival.
					MinimumSpeedSampleCount,
				MaximumSpeedSampleCount: policy.Arrival.
					MaximumSpeedSampleCount,
				MinimumGroundSpeedMPS: policy.Arrival.
					MinimumGroundSpeedMPS,
				SpeedUncertaintyMultiplier: policy.Arrival.
					SpeedUncertaintyMultiplier,
				MinimumArrivalInterval: policy.Arrival.
					MinimumArrivalInterval,
				MaximumEstimatedArrivalDuration: policy.Arrival.
					MaximumEstimatedArrivalDuration,
				MaximumExtrapolationConfidenceLoss: policy.Arrival.
					MaximumExtrapolationConfidenceLoss,
				ProjectionConfidenceWeight: policy.Arrival.
					ProjectionConfidenceWeight,
				DestinationConfidenceWeight: policy.Arrival.
					DestinationConfidenceWeight,
				SpeedStabilityWeight: policy.Arrival.
					SpeedStabilityWeight,
				MediumConfidenceMinimum: policy.Arrival.
					MediumConfidenceMinimum,
				HighConfidenceMinimum: policy.Arrival.
					HighConfidenceMinimum,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"Estimated Arrival evaluator",
				err,
			)
	}

	composer, err :=
		projectionproduction.New(
			projectionproduction.Config{
				HorizonPlanner:             horizonPlanner,
				KinematicProjector:         kinematic,
				HistoricalProjector:        historical,
				NeighborSelector:           selector,
				PatternConfidenceEvaluator: pattern,
				FreshnessEvaluator:         freshness,
				RouteFrequencyEvaluator:    frequency,
				ArrivalEstimator:           arrival,
				FreshnessLimitedPolicy: projectionproduction.
					LimitedEvidenceReject,
				RouteFrequencyLimitedPolicy: projectionproduction.
					LimitedEvidenceReject,
				DependencyFailurePolicy: projectionproduction.
					DependencyFailureFallbackToKinematic,
				ArrivalFailurePolicy: projectionproduction.
					ArrivalFailurePreserveProjection,
			},
		)
	if err != nil {
		return algorithmComponents{},
			componentError(
				"production projection composer",
				err,
			)
	}

	return algorithmComponents{
		composer: composer,
	}, nil
}

func componentError(
	component string,
	err error,
) error {
	return fmt.Errorf(
		"construct Projection Intelligence %s: %w",
		component,
		err,
	)
}
