package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	WeatherContextReadResultVersion = "weather-context-read-result-v1"

	weatherContextAsOfTimeQuery        = "as_of_time"
	weatherContextDurationSecondsQuery = "duration_seconds"
)

var (
	ErrWeatherContextNotFound = errors.New(
		"Weather Context result was not found",
	)
	ErrWeatherContextServiceUnavailable = errors.New(
		"Weather Context service is unavailable",
	)
	ErrWeatherContextInvalidRequest = errors.New(
		"Weather Context request is invalid",
	)

	errWeatherContextTrajectoryIDInvalid = errors.New(
		"trajectory identifier must be a valid UUID",
	)
	errWeatherContextAsOfTimeInvalid = errors.New(
		"as-of time must be a valid RFC 3339 timestamp",
	)
	errWeatherContextDurationInvalid = errors.New(
		"Weather Context duration must be a positive whole number of seconds",
	)
)

var weatherContextFingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

type WeatherContextReadRequest struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type WeatherContextReadResult struct {
	Version string

	Weather     weathercontract.Result
	Trust       weathertrust.Result
	Alignment   weatheralignment.Result
	Encounter   weatherencounter.Result
	Uncertainty weatheruncertainty.Result

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result WeatherContextReadResult) Clone() WeatherContextReadResult {
	cloned := result
	cloned.Weather = result.Weather.Clone()
	cloned.Trust = result.Trust.Clone()
	cloned.Alignment = result.Alignment.Clone()
	cloned.Encounter = result.Encounter.Clone()
	cloned.Uncertainty = result.Uncertainty.Clone()
	return cloned
}

func (result WeatherContextReadResult) Validate() error {
	if result.Version != WeatherContextReadResultVersion {
		return fmt.Errorf("Weather Context read-result version is invalid")
	}
	if !weatherContextFingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("Weather Context aggregate input fingerprint is invalid")
	}
	if result.GeneratedAt.IsZero() {
		return fmt.Errorf("Weather Context generated-at time is required")
	}

	weatherReport := weathercontract.Validate(result.Weather)
	if weatherReport.Status != weathercontract.ValidationStatusValid {
		return fmt.Errorf("Weather Feature Contract is invalid: %v", weatherReport.Issues)
	}
	if err := result.Trust.Validate(); err != nil {
		return fmt.Errorf("Weather Trust Gate result is invalid: %w", err)
	}
	if err := result.Alignment.Validate(); err != nil {
		return fmt.Errorf("Weather Alignment result is invalid: %w", err)
	}
	if err := result.Encounter.Validate(); err != nil {
		return fmt.Errorf("Weather Encounter result is invalid: %w", err)
	}
	if err := result.Uncertainty.Validate(); err != nil {
		return fmt.Errorf("Weather Uncertainty result is invalid: %w", err)
	}

	trajectoryID := strings.TrimSpace(result.Weather.TrajectoryID)
	if trajectoryID == "" ||
		strings.TrimSpace(result.Alignment.TrajectoryID) != trajectoryID ||
		strings.TrimSpace(result.Encounter.TrajectoryID) != trajectoryID ||
		strings.TrimSpace(result.Uncertainty.TrajectoryID) != trajectoryID {
		return fmt.Errorf("Weather Context trajectory identifiers are inconsistent")
	}

	asOfTime := result.Weather.AsOfTime.UTC()
	if !result.Trust.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Alignment.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Encounter.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Uncertainty.AsOfTime.UTC().Equal(asOfTime) {
		return fmt.Errorf("Weather Context as-of times are inconsistent")
	}

	for _, generatedAt := range []time.Time{
		result.Weather.GeneratedAt,
		result.Alignment.GeneratedAt,
		result.Encounter.GeneratedAt,
		result.Uncertainty.GeneratedAt,
	} {
		if result.GeneratedAt.Before(generatedAt) {
			return fmt.Errorf("Weather Context aggregate was generated before one of its inputs")
		}
	}

	return nil
}

type WeatherContextReader interface {
	GetWeatherContext(
		context.Context,
		WeatherContextReadRequest,
	) (WeatherContextReadResult, error)
}

type WeatherContextHandler struct {
	reader WeatherContextReader
}

func NewWeatherContextHandler(
	reader WeatherContextReader,
) *WeatherContextHandler {
	return &WeatherContextHandler{
		reader: reader,
	}
}

func (handler *WeatherContextHandler) GetByTrajectoryID(
	ctx *fiber.Ctx,
) error {
	if handler == nil || handler.reader == nil {
		return weatherContextUnavailable(ctx)
	}

	request, err := parseWeatherContextReadRequest(
		ctx.Params("id"),
		ctx.Query(weatherContextAsOfTimeQuery),
		ctx.Query(weatherContextDurationSecondsQuery),
	)
	if err != nil {
		return weatherContextRequestError(ctx, err)
	}

	result, err := handler.reader.GetWeatherContext(
		ctx.Context(),
		request,
	)
	if err != nil {
		return writeWeatherContextError(ctx, err)
	}
	if err := result.Validate(); err != nil {
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"WEATHER_CONTEXT_CONTRACT_INVALID",
			"Weather Context service returned an invalid result",
		)
	}

	return response.OK(
		ctx,
		dto.ToWeatherContextResponse(
			result.Weather,
			result.Trust,
			result.Alignment,
			result.Encounter,
			result.Uncertainty,
			result.InputFingerprint,
			result.GeneratedAt,
		),
	)
}

func parseWeatherContextReadRequest(
	trajectoryIDValue string,
	asOfTimeValue string,
	durationSecondsValue string,
) (WeatherContextReadRequest, error) {
	trajectoryID, err := parseWeatherContextTrajectoryID(
		trajectoryIDValue,
	)
	if err != nil {
		return WeatherContextReadRequest{}, err
	}

	asOfTime, err := parseWeatherContextAsOfTime(
		asOfTimeValue,
	)
	if err != nil {
		return WeatherContextReadRequest{}, err
	}

	duration, err := parseWeatherContextDuration(
		durationSecondsValue,
	)
	if err != nil {
		return WeatherContextReadRequest{}, err
	}

	return WeatherContextReadRequest{
		TrajectoryID:      trajectoryID,
		AsOfTime:          asOfTime,
		RequestedDuration: duration,
	}, nil
}

func parseWeatherContextTrajectoryID(
	value string,
) (string, error) {
	normalized := strings.TrimSpace(value)
	parsed, err := uuid.Parse(normalized)
	if normalized == "" || err != nil {
		return "", errWeatherContextTrajectoryIDInvalid
	}
	return strings.ToLower(parsed.String()), nil
}

func parseWeatherContextAsOfTime(
	value string,
) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{}, errWeatherContextAsOfTimeInvalid
	}
	parsed, err := time.Parse(
		time.RFC3339Nano,
		normalized,
	)
	if err != nil {
		return time.Time{}, errWeatherContextAsOfTimeInvalid
	}
	return parsed.UTC(), nil
}

func parseWeatherContextDuration(
	value string,
) (time.Duration, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, errWeatherContextDurationInvalid
	}
	seconds, err := strconv.ParseInt(
		normalized,
		10,
		64,
	)
	if err != nil ||
		seconds < 1 ||
		seconds > math.MaxInt64/int64(time.Second) {
		return 0, errWeatherContextDurationInvalid
	}
	return time.Duration(seconds) * time.Second, nil
}

func weatherContextRequestError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, errWeatherContextTrajectoryIDInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WEATHER_CONTEXT_TRAJECTORY_ID",
			"Weather Context trajectory identifier must be a valid UUID",
		)
	case errors.Is(err, errWeatherContextAsOfTimeInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WEATHER_CONTEXT_AS_OF_TIME",
			"Weather Context as-of time is required and must be a valid RFC 3339 timestamp",
		)
	case errors.Is(err, errWeatherContextDurationInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WEATHER_CONTEXT_DURATION",
			"Weather Context duration must be a positive whole number of seconds",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WEATHER_CONTEXT_REQUEST",
			"Weather Context request is invalid",
		)
	}
}

func weatherContextUnavailable(
	ctx *fiber.Ctx,
) error {
	return response.Error(
		ctx,
		fiber.StatusServiceUnavailable,
		"WEATHER_CONTEXT_SERVICE_UNAVAILABLE",
		"Weather Context service is unavailable",
	)
}

func writeWeatherContextError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return response.Error(
			ctx,
			fiber.StatusGatewayTimeout,
			"WEATHER_CONTEXT_TIMEOUT",
			"Weather Context request timed out",
		)
	case errors.Is(err, context.Canceled):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"WEATHER_CONTEXT_REQUEST_CANCELED",
			"Weather Context request was canceled",
		)
	case errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, ErrWeatherContextNotFound):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"WEATHER_CONTEXT_NOT_FOUND",
			"No matching trajectory was available for Weather Context",
		)
	case errors.Is(err, ErrWeatherContextServiceUnavailable):
		return weatherContextUnavailable(ctx)
	case errors.Is(err, ErrWeatherContextInvalidRequest):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WEATHER_CONTEXT_REQUEST",
			"Weather Context request is outside the configured policy",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"WEATHER_CONTEXT_LOAD_FAILED",
			"Failed to load Weather Context",
		)
	}
}
