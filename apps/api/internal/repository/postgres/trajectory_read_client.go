package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
// supplied PostgreSQL client. A transaction keeps trajectory metadata,
// segments, and coverage gaps inside its caller-owned snapshot. A pool is
// retained so public aggregate reads can create their own repeatable-read
// snapshot.
func NewTrajectoryReadRepository(
	client TrajectoryReadClient,
) *TrajectoryRepository {
	repository := &TrajectoryRepository{
		readClient: client,
	}

	if pool, ok := client.(*pgxpool.Pool); ok {
		repository.db = pool
	}

	return repository
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
