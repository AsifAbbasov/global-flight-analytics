package projectionread

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type trajectoryRepository interface {
	GetTrajectoryByID(
		context.Context,
		string,
	) (trajectory.FlightTrajectory, error)
}

type PostgresDataSourceConfig struct {
	Pool                 *pgxpool.Pool
	TrajectoryRepository trajectoryRepository
	Policy               DataSourcePolicy
}

type PostgresDataSource struct {
	client               postgresClient
	trajectoryRepository trajectoryRepository
	snapshotExecutor     snapshotExecutor
	policy               DataSourcePolicy
}

func NewPostgresDataSource(
	config PostgresDataSourceConfig,
) (*PostgresDataSource, error) {
	if config.Pool == nil {
		return nil, fmt.Errorf(
			"Projection Intelligence PostgreSQL pool is required",
		)
	}

	repository := config.TrajectoryRepository
	if repository == nil {
		repository = postgres.NewTrajectoryRepository(
			config.Pool,
		)
	}

	source, err := newPostgresDataSource(
		pgxPoolClient{
			pool: config.Pool,
		},
		repository,
		config.Policy,
	)
	if err != nil {
		return nil, err
	}
	source.snapshotExecutor = repeatableReadSnapshotExecutor{
		starter: pgxSnapshotTransactionStarter{
			pool: config.Pool,
		},
	}
	return source, nil
}

func newPostgresDataSource(
	client postgresClient,
	repository trajectoryRepository,
	policy DataSourcePolicy,
) (*PostgresDataSource, error) {
	if client == nil {
		return nil, fmt.Errorf(
			"Projection Intelligence PostgreSQL client is required",
		)
	}
	if repository == nil {
		return nil, fmt.Errorf(
			"Projection Intelligence trajectory repository is required",
		)
	}
	if policy.MaximumTrajectoryPointCount < 2 ||
		policy.MaximumHistoricalCandidateCount < 1 ||
		policy.HistoricalCandidateLookback <= 0 ||
		policy.RouteHistoryWindow <= 0 ||
		policy.RecentRouteWindow <= 0 ||
		policy.RecentRouteWindow >
			policy.RouteHistoryWindow ||
		policy.SourceName == "" {
		return nil, fmt.Errorf(
			"Projection Intelligence PostgreSQL data-source policy is invalid",
		)
	}

	return &PostgresDataSource{
		client:               client,
		trajectoryRepository: repository,
		snapshotExecutor: directSnapshotExecutor{
			client:     client,
			repository: repository,
		},
		policy: policy,
	}, nil
}
