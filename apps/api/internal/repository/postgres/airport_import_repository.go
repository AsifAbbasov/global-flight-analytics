package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
)

var ErrAirportImportRepositoryPoolRequired = errors.New(
	"airport import repository pool is required",
)

func (
	r *AirportRepository,
) UpsertImported(
	ctx context.Context,
	items []airport.ImportRecord,
) (int64, error) {
	if r == nil || r.pool == nil {
		return 0,
			ErrAirportImportRepositoryPoolRequired
	}

	if len(items) == 0 {
		return 0, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := r.pool.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return 0, fmt.Errorf(
			"begin airport import transaction: %w",
			err,
		)
	}

	committed := false
	defer func() {
		if !committed {
			rollbackRepositoryTransaction(tx)
		}
	}()

	if err := executeAirportImport(ctx, tx, items); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf(
			"commit airport import transaction: %w",
			err,
		)
	}
	committed = true

	return int64(len(items)), nil
}
