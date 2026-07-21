package postgres

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
)

const createAirportImportStagingTableQuery = `
	CREATE TEMP TABLE airport_import_staging (
		source_ident text NOT NULL,
		icao_code text,
		iata_code text,
		name text NOT NULL,
		city text,
		source_country_code text,
		latitude double precision NOT NULL,
		longitude double precision NOT NULL,
		elevation_ft integer,
		source_name text NOT NULL,
		last_synced_at timestamptz NOT NULL
	)
	ON COMMIT DROP;
`

const insertAirportImportStagingRecordQuery = `
	INSERT INTO airport_import_staging (
		source_ident,
		icao_code,
		iata_code,
		name,
		city,
		source_country_code,
		latitude,
		longitude,
		elevation_ft,
		source_name,
		last_synced_at
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11
	);
`

func createAirportImportStagingTable(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if _, err := tx.Exec(
		ctx,
		createAirportImportStagingTableQuery,
	); err != nil {
		return fmt.Errorf(
			"create airport import staging table: %w",
			err,
		)
	}

	return nil
}

func stageAirportImportRecords(
	ctx context.Context,
	tx pgx.Tx,
	items []airport.ImportRecord,
) error {
	batch := &pgx.Batch{}

	for _, item := range items {
		batch.Queue(
			insertAirportImportStagingRecordQuery,
			item.SourceIdent,
			item.ICAOCode,
			item.IATACode,
			item.Name,
			item.City,
			item.SourceCountryCode,
			item.Latitude,
			item.Longitude,
			item.ElevationFT,
			item.SourceName,
			item.LastSyncedAt,
		)
	}

	results := tx.SendBatch(ctx, batch)
	for index := range items {
		commandTag, err := results.Exec()
		if err != nil {
			_ = results.Close()
			return fmt.Errorf(
				"insert airport import staging record at index %d: %w",
				index,
				err,
			)
		}
		if commandTag.RowsAffected() != 1 {
			_ = results.Close()
			return fmt.Errorf(
				"insert airport import staging record at index %d: expected 1 affected row, got %d",
				index,
				commandTag.RowsAffected(),
			)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf(
			"close airport import staging batch: %w",
			err,
		)
	}

	return nil
}
