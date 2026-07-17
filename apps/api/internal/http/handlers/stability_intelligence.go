package handlers

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	stabilityIntelligenceAsOfTimesQuery       = "as_of_times"
	stabilityIntelligenceDurationSecondsQuery = "duration_seconds"
)

var (
	errStabilityTrajectoryIDInvalid = errors.New(
		"trajectory identifier must be a valid UUID",
	)
	errStabilityAsOfTimesInvalid = errors.New(
		"as-of times must contain two to eight strictly increasing RFC 3339 timestamps",
	)
	errStabilityDurationInvalid = errors.New(
		"duration must be a positive whole number of seconds",
	)
)

type StabilityIntelligenceReader interface {
	Get(
		context.Context,
		stabilityproduction.Request,
	) (stabilityproduction.Result, error)
}

type StabilityIntelligenceHandler struct {
	reader StabilityIntelligenceReader
}

func NewStabilityIntelligenceHandler(
	reader StabilityIntelligenceReader,
) *StabilityIntelligenceHandler {
	return &StabilityIntelligenceHandler{
		reader: reader,
	}
}

func (
	handler *StabilityIntelligenceHandler,
) GetByTrajectoryID(
	ctx *fiber.Ctx,
) error {
	if handler == nil ||
		handler.reader == nil {
		return stabilityIntelligenceUnavailable(
			ctx,
		)
	}

	request, err :=
		parseStabilityIntelligenceReadRequest(
			ctx.Params("id"),
			ctx.Query(
				stabilityIntelligenceAsOfTimesQuery,
			),
			ctx.Query(
				stabilityIntelligenceDurationSecondsQuery,
			),
		)
	if err != nil {
		return stabilityIntelligenceRequestError(
			ctx,
			err,
		)
	}

	result, err := handler.reader.Get(
		ctx.Context(),
		request,
	)
	if err != nil {
		return writeStabilityIntelligenceError(
			ctx,
			err,
		)
	}

	if err := result.Validate(); err != nil {
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"STABILITY_INTELLIGENCE_CONTRACT_INVALID",
			"Stability Intelligence service returned an invalid production result",
		)
	}

	return response.OK(
		ctx,
		dto.ToStabilityIntelligenceResponse(
			result,
		),
	)
}

func parseStabilityIntelligenceReadRequest(
	trajectoryIDValue string,
	asOfTimesValue string,
	durationSecondsValue string,
) (
	stabilityproduction.Request,
	error,
) {
	trajectoryID, err :=
		parseStabilityTrajectoryID(
			trajectoryIDValue,
		)
	if err != nil {
		return stabilityproduction.Request{},
			err
	}

	asOfTimes, err :=
		parseStabilityAsOfTimes(
			asOfTimesValue,
		)
	if err != nil {
		return stabilityproduction.Request{},
			err
	}

	duration, err :=
		parseStabilityDuration(
			durationSecondsValue,
		)
	if err != nil {
		return stabilityproduction.Request{},
			err
	}

	return stabilityproduction.Request{
		TrajectoryID:      trajectoryID,
		AsOfTimes:         asOfTimes,
		RequestedDuration: duration,
	}, nil
}

func parseStabilityTrajectoryID(
	value string,
) (string, error) {
	normalized := strings.TrimSpace(
		value,
	)
	parsed, err := uuid.Parse(
		normalized,
	)
	if normalized == "" ||
		err != nil {
		return "",
			errStabilityTrajectoryIDInvalid
	}

	return strings.ToLower(
		parsed.String(),
	), nil
}

func parseStabilityAsOfTimes(
	value string,
) ([]time.Time, error) {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" {
		return nil,
			errStabilityAsOfTimesInvalid
	}

	parts := strings.Split(
		normalized,
		",",
	)
	if len(parts) <
		stabilityproduction.
			MinimumAsOfTimeCount ||
		len(parts) >
			stabilityproduction.
				MaximumAsOfTimeCount {
		return nil,
			errStabilityAsOfTimesInvalid
	}

	result := make(
		[]time.Time,
		0,
		len(parts),
	)
	var previous time.Time
	for _, part := range parts {
		parsed, err := time.Parse(
			time.RFC3339Nano,
			strings.TrimSpace(
				part,
			),
		)
		if err != nil {
			return nil,
				errStabilityAsOfTimesInvalid
		}
		parsed = parsed.UTC()
		if !previous.IsZero() &&
			!parsed.After(previous) {
			return nil,
				errStabilityAsOfTimesInvalid
		}
		result = append(
			result,
			parsed,
		)
		previous = parsed
	}

	return result, nil
}

func parseStabilityDuration(
	value string,
) (time.Duration, error) {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" {
		return 0,
			errStabilityDurationInvalid
	}

	seconds, err := strconv.ParseInt(
		normalized,
		10,
		64,
	)
	if err != nil ||
		seconds < 1 ||
		seconds >
			math.MaxInt64/
				int64(time.Second) {
		return 0,
			errStabilityDurationInvalid
	}

	return time.Duration(seconds) *
			time.Second,
		nil
}

func stabilityIntelligenceRequestError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(
		err,
		errStabilityTrajectoryIDInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_STABILITY_TRAJECTORY_ID",
			"Stability Intelligence trajectory identifier must be a valid UUID",
		)

	case errors.Is(
		err,
		errStabilityAsOfTimesInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_STABILITY_AS_OF_TIMES",
			"Stability Intelligence as-of times must contain two to eight strictly increasing RFC 3339 timestamps",
		)

	case errors.Is(
		err,
		errStabilityDurationInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_STABILITY_DURATION",
			"Stability Intelligence duration must be a positive whole number of seconds",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_STABILITY_INTELLIGENCE_REQUEST",
			"Stability Intelligence request is invalid",
		)
	}
}

func stabilityIntelligenceUnavailable(
	ctx *fiber.Ctx,
) error {
	return response.Error(
		ctx,
		fiber.StatusServiceUnavailable,
		"STABILITY_INTELLIGENCE_SERVICE_UNAVAILABLE",
		"Stability Intelligence service is unavailable",
	)
}

func writeStabilityIntelligenceError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(
		err,
		context.DeadlineExceeded,
	):
		return response.Error(
			ctx,
			fiber.StatusGatewayTimeout,
			"STABILITY_INTELLIGENCE_TIMEOUT",
			"Stability Intelligence request timed out",
		)

	case errors.Is(
		err,
		context.Canceled,
	):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"STABILITY_INTELLIGENCE_REQUEST_CANCELED",
			"Stability Intelligence request was canceled",
		)

	case errors.Is(
		err,
		pgx.ErrNoRows,
	),
		errors.Is(
			err,
			stabilityproduction.
				ErrTrajectoryNotFound,
		):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"STABILITY_INTELLIGENCE_NOT_FOUND",
			"No matching trajectory was available for Stability Intelligence",
		)

	case errors.Is(
		err,
		stabilityproduction.
			ErrServiceUnavailable,
	):
		return stabilityIntelligenceUnavailable(
			ctx,
		)

	case errors.Is(
		err,
		stabilityproduction.
			ErrInvalidRequest,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_STABILITY_INTELLIGENCE_REQUEST",
			"Stability Intelligence request is outside the configured production policy",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"STABILITY_INTELLIGENCE_LOAD_FAILED",
			"Failed to load Stability Intelligence",
		)
	}
}
