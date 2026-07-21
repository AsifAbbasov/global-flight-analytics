package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const updateAirportsByICAOQuery = `
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

const updateAirportsBySourceIdentityQuery = `
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

const insertRemainingAirportsQuery = `
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

func updateAirportsByICAO(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if _, err := tx.Exec(ctx, updateAirportsByICAOQuery); err != nil {
		return fmt.Errorf("update airports by ICAO code: %w", err)
	}
	return nil
}

func updateAirportsBySourceIdentity(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if _, err := tx.Exec(
		ctx,
		updateAirportsBySourceIdentityQuery,
	); err != nil {
		return fmt.Errorf(
			"update airports by source identity: %w",
			err,
		)
	}
	return nil
}

func insertRemainingAirports(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if _, err := tx.Exec(ctx, insertRemainingAirportsQuery); err != nil {
		return fmt.Errorf("insert remaining airports: %w", err)
	}
	return nil
}
