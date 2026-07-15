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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	projectionIntelligenceAsOfTimeQuery        = "as_of_time"
	projectionIntelligenceDurationSecondsQuery = "duration_seconds"
)

var (
	ErrProjectionIntelligenceNotFound = errors.New(
		"Projection Intelligence result was not found",
	)
	ErrProjectionIntelligenceServiceUnavailable = errors.New(
		"Projection Intelligence service is unavailable",
	)
	ErrProjectionIntelligenceInvalidRequest = errors.New(
		"Projection Intelligence request is invalid",
	)

	errProjectionTrajectoryIDInvalid = errors.New(
		"trajectory identifier must be a valid UUID",
	)
	errProjectionAsOfTimeInvalid = errors.New(
		"as-of time must be a valid RFC 3339 timestamp",
	)
	errProjectionDurationInvalid = errors.New(
		"projection duration must be a positive whole number of seconds",
	)
)

type ProjectionIntelligenceReadRequest struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type ProjectionIntelligenceReader interface {
	GetProjectionIntelligence(
		context.Context,
		ProjectionIntelligenceReadRequest,
	) (projectionproduction.Result, error)
}

type ProjectionIntelligenceHandler struct {
	reader ProjectionIntelligenceReader
}

func NewProjectionIntelligenceHandler(
	reader ProjectionIntelligenceReader,
) *ProjectionIntelligenceHandler {
	return &ProjectionIntelligenceHandler{
		reader: reader,
	}
}

func (
	handler *ProjectionIntelligenceHandler,
) GetByTrajectoryID(
	ctx *fiber.Ctx,
) error {
	if handler == nil ||
		handler.reader == nil {
		return projectionIntelligenceUnavailable(
			ctx,
		)
	}

	request, err :=
		parseProjectionIntelligenceReadRequest(
			ctx.Params("id"),
			ctx.Query(
				projectionIntelligenceAsOfTimeQuery,
			),
			ctx.Query(
				projectionIntelligenceDurationSecondsQuery,
			),
		)
	if err != nil {
		return projectionIntelligenceRequestError(
			ctx,
			err,
		)
	}

	result, err :=
		handler.reader.GetProjectionIntelligence(
			ctx.Context(),
			request,
		)
	if err != nil {
		return writeProjectionIntelligenceError(
			ctx,
			err,
		)
	}

	if err := result.Validate(); err != nil {
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"PROJECTION_INTELLIGENCE_CONTRACT_INVALID",
			"Projection Intelligence service returned an invalid production result",
		)
	}

	return response.OK(
		ctx,
		dto.ToProjectionIntelligenceResponse(
			result,
		),
	)
}

func parseProjectionIntelligenceReadRequest(
	trajectoryIDValue string,
	asOfTimeValue string,
	durationSecondsValue string,
) (
	ProjectionIntelligenceReadRequest,
	error,
) {
	trajectoryID, err :=
		parseProjectionIntelligenceTrajectoryID(
			trajectoryIDValue,
		)
	if err != nil {
		return ProjectionIntelligenceReadRequest{},
			err
	}

	asOfTime, err :=
		parseProjectionIntelligenceAsOfTime(
			asOfTimeValue,
		)
	if err != nil {
		return ProjectionIntelligenceReadRequest{},
			err
	}

	duration, err :=
		parseProjectionIntelligenceDuration(
			durationSecondsValue,
		)
	if err != nil {
		return ProjectionIntelligenceReadRequest{},
			err
	}

	return ProjectionIntelligenceReadRequest{
		TrajectoryID:      trajectoryID,
		AsOfTime:          asOfTime,
		RequestedDuration: duration,
	}, nil
}

func parseProjectionIntelligenceTrajectoryID(
	value string,
) (string, error) {
	normalized := strings.TrimSpace(value)
	parsed, err := uuid.Parse(normalized)
	if normalized == "" ||
		err != nil {
		return "",
			errProjectionTrajectoryIDInvalid
	}

	return strings.ToLower(
		parsed.String(),
	), nil
}

func parseProjectionIntelligenceAsOfTime(
	value string,
) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{},
			errProjectionAsOfTimeInvalid
	}

	parsed, err := time.Parse(
		time.RFC3339Nano,
		normalized,
	)
	if err != nil {
		return time.Time{},
			errProjectionAsOfTimeInvalid
	}

	return parsed.UTC(), nil
}

func parseProjectionIntelligenceDuration(
	value string,
) (time.Duration, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0,
			errProjectionDurationInvalid
	}

	seconds, err := strconv.ParseInt(
		normalized,
		10,
		64,
	)
	if err != nil ||
		seconds < 1 ||
		seconds >
			math.MaxInt64/int64(time.Second) {
		return 0,
			errProjectionDurationInvalid
	}

	return time.Duration(seconds) *
			time.Second,
		nil
}

func projectionIntelligenceRequestError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(
		err,
		errProjectionTrajectoryIDInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_PROJECTION_TRAJECTORY_ID",
			"Projection Intelligence trajectory identifier must be a valid UUID",
		)

	case errors.Is(
		err,
		errProjectionAsOfTimeInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_PROJECTION_AS_OF_TIME",
			"Projection Intelligence as-of time is required and must be a valid RFC 3339 timestamp",
		)

	case errors.Is(
		err,
		errProjectionDurationInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_PROJECTION_DURATION",
			"Projection Intelligence duration must be a positive whole number of seconds",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_PROJECTION_INTELLIGENCE_REQUEST",
			"Projection Intelligence request is invalid",
		)
	}
}

func projectionIntelligenceUnavailable(
	ctx *fiber.Ctx,
) error {
	return response.Error(
		ctx,
		fiber.StatusServiceUnavailable,
		"PROJECTION_INTELLIGENCE_SERVICE_UNAVAILABLE",
		"Projection Intelligence service is unavailable",
	)
}

func writeProjectionIntelligenceError(
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
			"PROJECTION_INTELLIGENCE_TIMEOUT",
			"Projection Intelligence request timed out",
		)

	case errors.Is(
		err,
		context.Canceled,
	):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"PROJECTION_INTELLIGENCE_REQUEST_CANCELED",
			"Projection Intelligence request was canceled",
		)

	case errors.Is(
		err,
		pgx.ErrNoRows,
	),
		errors.Is(
			err,
			ErrProjectionIntelligenceNotFound,
		):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"PROJECTION_INTELLIGENCE_NOT_FOUND",
			"No matching trajectory was available for Projection Intelligence",
		)

	case errors.Is(
		err,
		ErrProjectionIntelligenceServiceUnavailable,
	):
		return projectionIntelligenceUnavailable(
			ctx,
		)

	case errors.Is(
		err,
		ErrProjectionIntelligenceInvalidRequest,
	),
		errors.Is(
			err,
			projectionhorizon.
				ErrRequestedDurationBelowMinimum,
		),
		errors.Is(
			err,
			projectionproduction.
				ErrTrajectoryIDRequired,
		):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_PROJECTION_INTELLIGENCE_REQUEST",
			"Projection Intelligence request is outside the configured projection policy",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"PROJECTION_INTELLIGENCE_LOAD_FAILED",
			"Failed to load Projection Intelligence",
		)
	}
}
