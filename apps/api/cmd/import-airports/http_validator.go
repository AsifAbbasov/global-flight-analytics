package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/sourcehttp"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/ourairports"
)

type HTTPValidatorRepository interface {
	Get(
		ctx context.Context,
		sourceName string,
		resourceURL string,
	) (sourcehttp.Validator, bool, error)

	Upsert(
		ctx context.Context,
		validator sourcehttp.Validator,
	) error
}

func loadConditionalRequest(
	ctx context.Context,
	repository HTTPValidatorRepository,
) (ourairports.ConditionalRequest, error) {
	validator, exists, err := repository.Get(
		ctx,
		ourairports.SourceName,
		ourairports.AirportsCSVURL,
	)
	if err != nil {
		return ourairports.ConditionalRequest{}, fmt.Errorf(
			"load OurAirports HTTP validator: %w",
			err,
		)
	}

	if !exists {
		return ourairports.ConditionalRequest{}, nil
	}

	return ourairports.ConditionalRequest{
		ETag:         validator.ETag,
		LastModified: validator.LastModified,
	}, nil
}

func persistHTTPValidatorIfChanged(
	ctx context.Context,
	repository HTTPValidatorRepository,
	previousRequest ourairports.ConditionalRequest,
	result ourairports.LoadResult,
) (bool, error) {
	if sameHTTPValidators(previousRequest, result) {
		return false, nil
	}

	if err := persistHTTPValidator(ctx, repository, result); err != nil {
		return false, err
	}

	return true, nil
}

func sameHTTPValidators(
	previousRequest ourairports.ConditionalRequest,
	result ourairports.LoadResult,
) bool {
	return strings.TrimSpace(previousRequest.ETag) ==
		strings.TrimSpace(result.ETag) &&
		strings.TrimSpace(previousRequest.LastModified) ==
			strings.TrimSpace(result.LastModified)
}

func persistHTTPValidator(
	ctx context.Context,
	repository HTTPValidatorRepository,
	result ourairports.LoadResult,
) error {
	validator := sourcehttp.Validator{
		SourceName:   ourairports.SourceName,
		ResourceURL:  ourairports.AirportsCSVURL,
		ETag:         result.ETag,
		LastModified: result.LastModified,
		ObservedAt:   result.CheckedAt,
	}

	if !validator.HasValidators() {
		return nil
	}

	if err := repository.Upsert(ctx, validator); err != nil {
		return fmt.Errorf(
			"persist OurAirports HTTP validator: %w",
			err,
		)
	}

	return nil
}
