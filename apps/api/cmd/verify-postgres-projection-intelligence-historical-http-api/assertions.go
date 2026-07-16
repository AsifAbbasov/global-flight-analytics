package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/gofiber/fiber/v2"
)

type runtimeReader struct {
	service *projectionread.Service
}

func (
	reader runtimeReader,
) GetProjectionIntelligence(
	ctx context.Context,
	request handlers.ProjectionIntelligenceReadRequest,
) (
	projectionproduction.Result,
	error,
) {
	if reader.service == nil {
		return projectionproduction.Result{},
			handlers.
				ErrProjectionIntelligenceServiceUnavailable
	}

	result, err := reader.service.Get(
		ctx,
		projectionread.Request{
			TrajectoryID:      request.TrajectoryID,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			projectionread.ErrTrajectoryNotFound,
		):
			return projectionproduction.Result{},
				handlers.
					ErrProjectionIntelligenceNotFound

		case errors.Is(
			err,
			projectionread.ErrServiceUnavailable,
		):
			return projectionproduction.Result{},
				handlers.
					ErrProjectionIntelligenceServiceUnavailable

		case errors.Is(
			err,
			projectionread.ErrInvalidRequest,
		):
			return projectionproduction.Result{},
				handlers.
					ErrProjectionIntelligenceInvalidRequest

		default:
			return projectionproduction.Result{},
				err
		}
	}

	return result.Clone(), nil
}

func buildRuntimeApp(
	service *projectionread.Service,
) (*fiber.App, error) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	if err := server.RegisterProjectionIntelligenceReadRoute(
		v1,
		runtimeReader{
			service: service,
		},
	); err != nil {
		return nil,
			fmt.Errorf(
				"register Projection Intelligence HTTP route: %w",
				err,
			)
	}

	return app, nil
}

func verifyHistoricalEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
) (
	response.SuccessResponse[dto.ProjectionIntelligenceResponse],
	error,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		projectionRequestURL(
			verificationFlights[0].
				TrajectoryID,
			schedule.AsOfTime,
			verificationDuration,
		),
		nil,
	)
	httpResponse, err := app.Test(request)
	if err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"execute historical Projection Intelligence request: %w",
				err,
			)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(
			httpResponse.Body,
		)
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"status = %d, want %d, body = %s",
				httpResponse.StatusCode,
				fiber.StatusOK,
				body,
			)
	}

	var payload response.SuccessResponse[dto.ProjectionIntelligenceResponse]
	if err := json.NewDecoder(
		httpResponse.Body,
	).Decode(&payload); err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"decode historical Projection Intelligence response: %w",
				err,
			)
	}

	if err := validateHistoricalPayload(
		payload,
		schedule,
	); err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			err
	}

	return payload, nil
}

func validateHistoricalPayload(
	payload response.SuccessResponse[dto.ProjectionIntelligenceResponse],
	schedule verificationSchedule,
) error {
	data := payload.Data

	if !payload.Success {
		return fmt.Errorf(
			"success response flag is false",
		)
	}
	if data.Version !=
		projectionproduction.Version {
		return fmt.Errorf(
			"production version = %q, want %q",
			data.Version,
			projectionproduction.Version,
		)
	}
	if data.Strategy !=
		string(
			projectionproduction.
				StrategyHistoricalNeighbor,
		) {
		return fmt.Errorf(
			"strategy = %q, want historical neighbor continuation",
			data.Strategy,
		)
	}
	if data.FallbackReason != "" {
		return fmt.Errorf(
			"historical result unexpectedly contains fallback reason %q",
			data.FallbackReason,
		)
	}
	if data.Projection.Method.Name !=
		projectioncontinuation.MethodName {
		return fmt.Errorf(
			"projection method = %q, want %q",
			data.Projection.Method.Name,
			projectioncontinuation.MethodName,
		)
	}
	if data.Projection.TrajectoryID !=
		verificationFlights[0].TrajectoryID {
		return fmt.Errorf(
			"trajectory ID = %q, want %q",
			data.Projection.TrajectoryID,
			verificationFlights[0].
				TrajectoryID,
		)
	}
	if !data.Projection.Horizon.AsOfTime.Equal(
		schedule.AsOfTime,
	) ||
		data.Projection.Horizon.DurationSeconds !=
			int64(
				verificationDuration/
					time.Second,
			) {
		return fmt.Errorf(
			"unexpected projection horizon: %#v",
			data.Projection.Horizon,
		)
	}
	if len(data.Projection.Points) != 6 {
		return fmt.Errorf(
			"forecast point count = %d, want 6",
			len(data.Projection.Points),
		)
	}

	lastPointIndex :=
		len(data.Projection.Points) - 1
	if !data.Projection.Points[0].
		ForecastTime.Equal(
		schedule.AsOfTime.Add(
			30*time.Second,
		),
	) ||
		!data.Projection.Points[lastPointIndex].
			ForecastTime.Equal(
			schedule.AsOfTime.Add(
				verificationDuration,
			),
		) {
		return fmt.Errorf(
			"historical forecast timestamps do not cover the configured horizon",
		)
	}
	if data.Projection.Points[0].
		Uncertainty.HorizontalRadiusM <= 0 ||
		data.Projection.Points[lastPointIndex].
			Uncertainty.HorizontalRadiusM <= 0 {
		return fmt.Errorf(
			"historical projection uncertainty is unavailable",
		)
	}
	if data.Projection.Confidence.Score <= 0 ||
		data.Projection.Confidence.Level ==
			"none" {
		return fmt.Errorf(
			"historical projection confidence is unavailable: %#v",
			data.Projection.Confidence,
		)
	}
	if data.Projection.ScopeGuard !=
		"research_only_not_for_operational_use" {
		return fmt.Errorf(
			"unexpected scope guard: %q",
			data.Projection.ScopeGuard,
		)
	}

	if data.Evidence.NeighborSelection == nil ||
		data.Evidence.NeighborSelection.Status !=
			"complete" ||
		len(
			data.Evidence.NeighborSelection.
				Neighbors,
		) < 2 {
		return fmt.Errorf(
			"historical neighbor selection is not complete: %#v",
			data.Evidence.NeighborSelection,
		)
	}
	if data.Evidence.PatternConfidence == nil ||
		!data.Evidence.PatternConfidence.Usable ||
		data.Evidence.PatternConfidence.Score <= 0 {
		return fmt.Errorf(
			"pattern confidence did not authorize historical continuation: %#v",
			data.Evidence.PatternConfidence,
		)
	}
	if data.Evidence.Freshness == nil ||
		!data.Evidence.Freshness.Usable ||
		data.Evidence.Freshness.Decision !=
			"allowed" {
		return fmt.Errorf(
			"freshness guard did not allow historical continuation: %#v",
			data.Evidence.Freshness,
		)
	}
	if data.Evidence.RouteFrequency == nil ||
		!data.Evidence.RouteFrequency.Usable ||
		data.Evidence.RouteFrequency.Decision !=
			"allowed" ||
		data.Evidence.RouteFrequency.
			ObservationCount < 5 ||
		data.Evidence.RouteFrequency.
			DistinctDayCount < 5 {
		return fmt.Errorf(
			"route-frequency guard did not allow historical continuation: %#v",
			data.Evidence.RouteFrequency,
		)
	}
	if !hasNotice(
		data.Notices,
		"historical_neighbor_continuation_authorized",
	) {
		return fmt.Errorf(
			"historical continuation authorization notice is missing",
		)
	}
	if hasNotice(
		data.Notices,
		"historical_projection_failed",
	) ||
		hasNotice(
			data.Notices,
			"historical_projector_internal_fallback",
		) {
		return fmt.Errorf(
			"historical path unexpectedly used a kinematic fallback: %#v",
			data.Notices,
		)
	}

	if data.ArrivalStatus !=
		string(
			projectionproduction.
				ArrivalStatusAttached,
		) ||
		data.Projection.Arrival == nil ||
		data.Projection.Arrival.
			AirportICAOCode != "ZBBB" {
		return fmt.Errorf(
			"Estimated Arrival was not attached to the synthetic destination: status=%q arrival=%#v",
			data.ArrivalStatus,
			data.Projection.Arrival,
		)
	}
	if data.Projection.Arrival.
		EstimatedTime.Before(
		schedule.AsOfTime,
	) {
		return fmt.Errorf(
			"Estimated Arrival precedes the analytical time",
		)
	}

	if !strings.HasPrefix(
		data.InputFingerprint,
		"sha256:",
	) ||
		!strings.HasPrefix(
			data.Projection.Provenance.
				InputFingerprint,
			"sha256:",
		) {
		return fmt.Errorf(
			"deterministic fingerprints are missing",
		)
	}

	return nil
}

func projectionRequestURL(
	trajectoryID string,
	asOfTime time.Time,
	duration time.Duration,
) string {
	values := url.Values{}
	values.Set(
		"as_of_time",
		asOfTime.UTC().Format(
			time.RFC3339Nano,
		),
	)
	values.Set(
		"duration_seconds",
		fmt.Sprintf(
			"%d",
			int64(duration/time.Second),
		),
	)

	return "/api/v1/trajectories/" +
		trajectoryID +
		"/projection-intelligence?" +
		values.Encode()
}

func hasNotice(
	items []dto.ProjectionIntelligenceNotice,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}
