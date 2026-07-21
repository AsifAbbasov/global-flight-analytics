package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
)

var (
	ErrIngestionRunCountsInvalid = errors.New(
		"ingestion run processed counts exceed received records",
	)
	ErrIngestionRunErrorMessageInvalid = errors.New(
		"ingestion run error message is inconsistent with terminal status",
	)
)

func validateIngestionRunCompletion(
	status ingestionrun.Status,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) (string, error) {
	if recordsReceived < 0 ||
		recordsInserted < 0 ||
		recordsUpdated < 0 ||
		recordsInserted > recordsReceived ||
		recordsUpdated > recordsReceived-recordsInserted {
		return "", ErrIngestionRunCountsInvalid
	}

	normalizedErrorMessage := strings.TrimSpace(errorMessage)

	switch status {
	case ingestionrun.StatusSuccess:
		if normalizedErrorMessage != "" {
			return "", ErrIngestionRunErrorMessageInvalid
		}

	case ingestionrun.StatusFailed, ingestionrun.StatusPartial:
		if normalizedErrorMessage == "" {
			return "", ErrIngestionRunErrorMessageInvalid
		}

	default:
		return "", fmt.Errorf(
			"%w: unsupported terminal status %q",
			ErrIngestionRunErrorMessageInvalid,
			status,
		)
	}

	return normalizedErrorMessage, nil
}
