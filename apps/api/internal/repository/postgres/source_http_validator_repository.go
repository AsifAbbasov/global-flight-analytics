package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/sourcehttp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSourceHTTPValidatorRepositoryPoolRequired = errors.New(
	"source HTTP validator repository pool is required",
)

type SourceHTTPValidatorRepository struct {
	pool *pgxpool.Pool
}

func NewSourceHTTPValidatorRepository(
	pool *pgxpool.Pool,
) *SourceHTTPValidatorRepository {
	return &SourceHTTPValidatorRepository{
		pool: pool,
	}
}

func (repository *SourceHTTPValidatorRepository) Get(
	ctx context.Context,
	sourceName string,
	resourceURL string,
) (sourcehttp.Validator, bool, error) {
	if repository == nil || repository.pool == nil {
		return sourcehttp.Validator{},
			false,
			ErrSourceHTTPValidatorRepositoryPoolRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			source_name,
			resource_url,
			COALESCE(etag, ''),
			COALESCE(last_modified, ''),
			observed_at
		FROM source_http_validators
		WHERE source_name = $1
			AND resource_url = $2;
	`

	var validator sourcehttp.Validator

	err := repository.pool.QueryRow(
		ctx,
		query,
		sourceName,
		resourceURL,
	).Scan(
		&validator.SourceName,
		&validator.ResourceURL,
		&validator.ETag,
		&validator.LastModified,
		&validator.ObservedAt,
	)
	if err != nil {
		if errors.Is(
			err,
			pgx.ErrNoRows,
		) {
			return sourcehttp.Validator{},
				false,
				nil
		}

		return sourcehttp.Validator{},
			false,
			fmt.Errorf(
				"get source HTTP validator: %w",
				err,
			)
	}

	return validator,
		true,
		nil
}

func (repository *SourceHTTPValidatorRepository) Upsert(
	ctx context.Context,
	validator sourcehttp.Validator,
) error {
	if repository == nil || repository.pool == nil {
		return ErrSourceHTTPValidatorRepositoryPoolRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		INSERT INTO source_http_validators (
			source_name,
			resource_url,
			etag,
			last_modified,
			observed_at
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			$5
		)
		ON CONFLICT (
			source_name,
			resource_url
		)
		DO UPDATE SET
			etag = EXCLUDED.etag,
			last_modified = EXCLUDED.last_modified,
			observed_at = EXCLUDED.observed_at,
			updated_at = now();
	`

	_, err := repository.pool.Exec(
		ctx,
		query,
		validator.SourceName,
		validator.ResourceURL,
		validator.ETag,
		validator.LastModified,
		validator.ObservedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"upsert source HTTP validator: %w",
			err,
		)
	}

	return nil
}
