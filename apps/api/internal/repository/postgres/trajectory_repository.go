package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type TrajectoryRepository struct {
	db         *pgxpool.Pool
	readClient TrajectoryReadClient
}

func NewTrajectoryRepository(
	db *pgxpool.Pool,
) *TrajectoryRepository {
	return &TrajectoryRepository{
		db:         db,
		readClient: db,
	}
}
