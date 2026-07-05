package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/sourcehttp"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/ourairports"
)

type recordingHTTPValidatorRepository struct {
	getValidator sourcehttp.Validator
	getExists    bool
	getErr       error

	upsertedValidators []sourcehttp.Validator
	upsertErr          error
}

func (repository *recordingHTTPValidatorRepository) Get(
	ctx context.Context,
	sourceName string,
	resourceURL string,
) (sourcehttp.Validator, bool, error) {
	if repository.getErr != nil {
		return sourcehttp.Validator{}, false, repository.getErr
	}

	return repository.getValidator, repository.getExists, nil
}

func (repository *recordingHTTPValidatorRepository) Upsert(
	ctx context.Context,
	validator sourcehttp.Validator,
) error {
	if repository.upsertErr != nil {
		return repository.upsertErr
	}

	repository.upsertedValidators = append(
		repository.upsertedValidators,
		validator,
	)

	return nil
}

func TestLoadConditionalRequestReturnsEmptyRequestWhenValidatorDoesNotExist(
	t *testing.T,
) {
	repository := &recordingHTTPValidatorRepository{
		getExists: false,
	}

	request, err := loadConditionalRequest(
		context.Background(),
		repository,
	)
	if err != nil {
		t.Fatalf(
			"load conditional request without stored validator: %v",
			err,
		)
	}

	if request.ETag != "" {
		t.Fatalf(
			"expected empty ETag, got %q",
			request.ETag,
		)
	}

	if request.LastModified != "" {
		t.Fatalf(
			"expected empty Last-Modified value, got %q",
			request.LastModified,
		)
	}
}

func TestLoadConditionalRequestReturnsStoredValidatorState(
	t *testing.T,
) {
	storedValidator := sourcehttp.Validator{
		SourceName:   ourairports.SourceName,
		ResourceURL:  ourairports.AirportsCSVURL,
		ETag:         `"validator-a"`,
		LastModified: "Sun, 05 Jul 2026 01:53:55 GMT",
	}

	repository := &recordingHTTPValidatorRepository{
		getValidator: storedValidator,
		getExists:    true,
	}

	request, err := loadConditionalRequest(
		context.Background(),
		repository,
	)
	if err != nil {
		t.Fatalf(
			"load conditional request from stored validator: %v",
			err,
		)
	}

	if request.ETag != storedValidator.ETag {
		t.Fatalf(
			"unexpected ETag: got %q, want %q",
			request.ETag,
			storedValidator.ETag,
		)
	}

	if request.LastModified != storedValidator.LastModified {
		t.Fatalf(
			"unexpected Last-Modified value: got %q, want %q",
			request.LastModified,
			storedValidator.LastModified,
		)
	}
}

func TestLoadConditionalRequestPropagatesRepositoryError(
	t *testing.T,
) {
	expectedError := errors.New("validator repository unavailable")

	repository := &recordingHTTPValidatorRepository{
		getErr: expectedError,
	}

	_, err := loadConditionalRequest(
		context.Background(),
		repository,
	)
	if err == nil {
		t.Fatal(
			"expected repository error",
		)
	}

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"expected wrapped repository error, got %v",
			err,
		)
	}
}

func TestPersistHTTPValidatorIfChangedSkipsUnchangedValidatorState(
	t *testing.T,
) {
	repository := &recordingHTTPValidatorRepository{}

	previousRequest := ourairports.ConditionalRequest{
		ETag:         `"validator-a"`,
		LastModified: "Sun, 05 Jul 2026 01:53:55 GMT",
	}

	result := ourairports.LoadResult{
		CheckedAt: time.Date(
			2026,
			time.July,
			5,
			11,
			12,
			44,
			0,
			time.UTC,
		),
		ETag:         `"validator-a"`,
		LastModified: "Sun, 05 Jul 2026 01:53:55 GMT",
		NotModified:  true,
	}

	written, err := persistHTTPValidatorIfChanged(
		context.Background(),
		repository,
		previousRequest,
		result,
	)
	if err != nil {
		t.Fatalf(
			"persist unchanged HTTP validator state: %v",
			err,
		)
	}

	if written {
		t.Fatal(
			"expected unchanged HTTP validator state write to be skipped",
		)
	}

	if len(repository.upsertedValidators) != 0 {
		t.Fatalf(
			"expected zero validator writes, got %d",
			len(repository.upsertedValidators),
		)
	}
}

func TestPersistHTTPValidatorIfChangedStoresChangedValidatorState(
	t *testing.T,
) {
	repository := &recordingHTTPValidatorRepository{}

	previousRequest := ourairports.ConditionalRequest{
		ETag:         `"validator-a"`,
		LastModified: "Sun, 05 Jul 2026 01:53:55 GMT",
	}

	checkedAt := time.Date(
		2026,
		time.July,
		5,
		11,
		15,
		0,
		0,
		time.UTC,
	)

	result := ourairports.LoadResult{
		CheckedAt:    checkedAt,
		ETag:         `"validator-b"`,
		LastModified: "Sun, 05 Jul 2026 02:10:00 GMT",
		NotModified:  true,
	}

	written, err := persistHTTPValidatorIfChanged(
		context.Background(),
		repository,
		previousRequest,
		result,
	)
	if err != nil {
		t.Fatalf(
			"persist changed HTTP validator state: %v",
			err,
		)
	}

	if !written {
		t.Fatal(
			"expected changed HTTP validator state to be written",
		)
	}

	if len(repository.upsertedValidators) != 1 {
		t.Fatalf(
			"expected one validator write, got %d",
			len(repository.upsertedValidators),
		)
	}

	storedValidator := repository.upsertedValidators[0]

	if storedValidator.SourceName != ourairports.SourceName {
		t.Fatalf(
			"unexpected source name: %q",
			storedValidator.SourceName,
		)
	}

	if storedValidator.ResourceURL != ourairports.AirportsCSVURL {
		t.Fatalf(
			"unexpected resource URL: %q",
			storedValidator.ResourceURL,
		)
	}

	if storedValidator.ETag != result.ETag {
		t.Fatalf(
			"unexpected ETag: %q",
			storedValidator.ETag,
		)
	}

	if storedValidator.LastModified != result.LastModified {
		t.Fatalf(
			"unexpected Last-Modified value: %q",
			storedValidator.LastModified,
		)
	}

	if !storedValidator.ObservedAt.Equal(checkedAt) {
		t.Fatalf(
			"unexpected observed time: got %s, want %s",
			storedValidator.ObservedAt,
			checkedAt,
		)
	}
}

func TestPersistHTTPValidatorIfChangedPropagatesRepositoryError(
	t *testing.T,
) {
	expectedError := errors.New("validator repository unavailable")

	repository := &recordingHTTPValidatorRepository{
		upsertErr: expectedError,
	}

	previousRequest := ourairports.ConditionalRequest{
		ETag: `"validator-a"`,
	}

	result := ourairports.LoadResult{
		CheckedAt: time.Date(
			2026,
			time.July,
			5,
			11,
			20,
			0,
			0,
			time.UTC,
		),
		ETag:        `"validator-b"`,
		NotModified: true,
	}

	written, err := persistHTTPValidatorIfChanged(
		context.Background(),
		repository,
		previousRequest,
		result,
	)

	if written {
		t.Fatal(
			"expected failed validator write not to be reported as written",
		)
	}

	if err == nil {
		t.Fatal(
			"expected repository error",
		)
	}

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"expected wrapped repository error, got %v",
			err,
		)
	}
}
