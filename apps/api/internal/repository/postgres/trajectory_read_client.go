package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TrajectoryReadClient is the minimal PostgreSQL read contract used by the
// trajectory repository. Both a connection pool and a transaction satisfy it.
type TrajectoryReadClient interface {
	QueryRow(
		context.Context,
		string,
		...any,
	) pgx.Row
	Query(
		context.Context,
		string,
		...any,
	) (pgx.Rows, error)
}

// NewTrajectoryReadRepository creates a read-only repository bound to the
// supplied PostgreSQL client. Passing a pgx transaction keeps trajectory
// metadata, segments, and coverage gaps inside the same database snapshot.
func NewTrajectoryReadRepository(
	client TrajectoryReadClient,
) *TrajectoryRepository {
	return &TrajectoryRepository{
		readClient: client,
	}
}

func (
	repository *TrajectoryRepository,
) trajectoryReadExecutor() TrajectoryReadClient {
	if repository == nil {
		return nil
	}
	if repository.readClient != nil {
		return repository.readClient
	}
	return repository.db
}
