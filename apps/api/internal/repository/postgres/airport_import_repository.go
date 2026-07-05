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

func (r *AirportRepository) UpsertImported(
	ctx context.Context,
	items []airport.ImportRecord,
) (int64, error) {
	if r == nil || r.pool == nil {
		return 0, ErrAirportImportRepositoryPoolRequired
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
			_ = tx.Rollback(ctx)
		}
	}()

	const createStagingTableQuery = `
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

	if _, err := tx.Exec(
		ctx,
		createStagingTableQuery,
	); err != nil {
		return 0, fmt.Errorf(
			"create airport import staging table: %w",
			err,
		)
	}

	copiedCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{
			"airport_import_staging",
		},
		[]string{
			"source_ident",
			"icao_code",
			"iata_code",
			"name",
			"city",
			"source_country_code",
			"latitude",
			"longitude",
			"elevation_ft",
			"source_name",
			"last_synced_at",
		},
		pgx.CopyFromSlice(
			len(items),
			func(index int) ([]any, error) {
				item := items[index]

				return []any{
					item.SourceIdent,
					item.ICAOCode,
					item.IATACode,
					item.Name,
					item.City,
					item.SourceCountryCode,
					item.Latitude,
					item.Longitude,
					nullableAirportElevationFT(
						item.ElevationFT,
					),
					item.SourceName,
					item.LastSyncedAt,
				}, nil
			},
		),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"copy airports into staging table: %w",
			err,
		)
	}

	if copiedCount != int64(len(items)) {
		return 0, fmt.Errorf(
			"copy airports into staging table: expected %d rows, copied %d",
			len(items),
			copiedCount,
		)
	}

	const updateByICAOQuery = `
		UPDATE airports AS target
		SET
			source_ident = staging.source_ident,
			iata_code = NULLIF(
				staging.iata_code,
				''
			),
			name = staging.name,
			city = NULLIF(
				staging.city,
				''
			),
			country_id = COALESCE(
				country.id,
				target.country_id
			),
			source_country_code = NULLIF(
				staging.source_country_code,
				''
			),
			latitude = staging.latitude,
			longitude = staging.longitude,
			elevation_ft = staging.elevation_ft,
			source_name = staging.source_name,
			last_synced_at = staging.last_synced_at,
			updated_at = now()
		FROM airport_import_staging AS staging
		LEFT JOIN countries AS country
			ON country.iso2 = NULLIF(
				staging.source_country_code,
				''
			)
		WHERE target.icao_code = NULLIF(
			staging.icao_code,
			''
		)
			AND NULLIF(
				staging.icao_code,
				''
			) IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM airports AS source_match
				WHERE source_match.source_name = staging.source_name
					AND source_match.source_ident = staging.source_ident
					AND source_match.id <> target.id
			);
	`

	if _, err := tx.Exec(
		ctx,
		updateByICAOQuery,
	); err != nil {
		return 0, fmt.Errorf(
			"update airports by ICAO code: %w",
			err,
		)
	}

	const updateBySourceIdentityQuery = `
		UPDATE airports AS target
		SET
			icao_code = NULLIF(
				staging.icao_code,
				''
			),
			iata_code = NULLIF(
				staging.iata_code,
				''
			),
			name = staging.name,
			city = NULLIF(
				staging.city,
				''
			),
			country_id = COALESCE(
				country.id,
				target.country_id
			),
			source_country_code = NULLIF(
				staging.source_country_code,
				''
			),
			latitude = staging.latitude,
			longitude = staging.longitude,
			elevation_ft = staging.elevation_ft,
			last_synced_at = staging.last_synced_at,
			updated_at = now()
		FROM airport_import_staging AS staging
		LEFT JOIN countries AS country
			ON country.iso2 = NULLIF(
				staging.source_country_code,
				''
			)
		WHERE target.source_name = staging.source_name
			AND target.source_ident = staging.source_ident;
	`

	if _, err := tx.Exec(
		ctx,
		updateBySourceIdentityQuery,
	); err != nil {
		return 0, fmt.Errorf(
			"update airports by source identity: %w",
			err,
		)
	}

	const insertRemainingQuery = `
		INSERT INTO airports (
			source_ident,
			icao_code,
			iata_code,
			name,
			city,
			country_id,
			source_country_code,
			latitude,
			longitude,
			elevation_ft,
			source_name,
			last_synced_at
		)
		SELECT
			staging.source_ident,
			NULLIF(
				staging.icao_code,
				''
			),
			NULLIF(
				staging.iata_code,
				''
			),
			staging.name,
			NULLIF(
				staging.city,
				''
			),
			country.id,
			NULLIF(
				staging.source_country_code,
				''
			),
			staging.latitude,
			staging.longitude,
			staging.elevation_ft,
			staging.source_name,
			staging.last_synced_at
		FROM airport_import_staging AS staging
		LEFT JOIN countries AS country
			ON country.iso2 = NULLIF(
				staging.source_country_code,
				''
			)
		WHERE NOT EXISTS (
			SELECT 1
			FROM airports AS source_match
			WHERE source_match.source_name = staging.source_name
				AND source_match.source_ident = staging.source_ident
		)
			AND (
				NULLIF(
					staging.icao_code,
					''
				) IS NULL
				OR NOT EXISTS (
					SELECT 1
					FROM airports AS icao_match
					WHERE icao_match.icao_code = NULLIF(
						staging.icao_code,
						''
					)
				)
			);
	`

	if _, err := tx.Exec(
		ctx,
		insertRemainingQuery,
	); err != nil {
		return 0, fmt.Errorf(
			"insert remaining airports: %w",
			err,
		)
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

func nullableAirportElevationFT(
	value *int,
) any {
	if value == nil {
		return nil
	}

	return *value
}
