package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
)

func executeAirportImport(
	ctx context.Context,
	tx pgx.Tx,
	items []airport.ImportRecord,
) error {
	if err := createAirportImportStagingTable(ctx, tx); err != nil {
		return err
	}
	if err := stageAirportImportRecords(ctx, tx, items); err != nil {
		return err
	}
	if err := updateAirportsByICAO(ctx, tx); err != nil {
		return err
	}
	if err := updateAirportsBySourceIdentity(ctx, tx); err != nil {
		return err
	}
	if err := insertRemainingAirports(ctx, tx); err != nil {
		return err
	}

	return nil
}
