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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/gofiber/fiber/v2"
)

type productionProjectionReader interface {
	Get(
		context.Context,
		projectionread.Request,
	) (projectionproduction.Result, error)
}

type runtimeReader struct {
	service productionProjectionReader
	timeout time.Duration
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
	if reader.service == nil ||
		reader.timeout <= 0 {
		return projectionproduction.Result{},
			handlers.
				ErrProjectionIntelligenceServiceUnavailable
	}

	operationContext, cancel := context.WithTimeout(
		ctx,
		reader.timeout,
	)
	defer cancel()

	result, err := reader.service.Get(
		operationContext,
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
	service productionProjectionReader,
) (*fiber.App, error) {
	if service == nil {
		return nil,
			fmt.Errorf(
				"Projection Intelligence service is required",
			)
	}
	if historicalReadTimeout <= 0 ||
		historicalHTTPTestTimeout <=
			historicalReadTimeout {
		return nil,
			fmt.Errorf(
				"historical runtime timeout policy is invalid",
			)
	}

	app := fiber.New()
	v1 := app.Group("/api/v1")

	if err := server.RegisterProjectionIntelligenceReadRoute(
		v1,
		runtimeReader{
			service: service,
			timeout: historicalReadTimeout,
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

func verifyHistoricalService(
	ctx context.Context,
	service productionProjectionReader,
	schedule verificationSchedule,
) (projectionproduction.Result, error) {
	if service == nil {
		return projectionproduction.Result{},
			fmt.Errorf(
				"Projection Intelligence service is required",
			)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	operationContext, cancel := context.WithTimeout(
		ctx,
		historicalReadTimeout,
	)
	defer cancel()

	result, err := service.Get(
		operationContext,
		projectionread.Request{
			TrajectoryID: verificationFlights[0].
				TrajectoryID,
			AsOfTime:          schedule.AsOfTime,
			RequestedDuration: verificationDuration,
		},
	)
	if err != nil {
		return projectionproduction.Result{},
			fmt.Errorf(
				"execute production Projection Intelligence service: %w",
				err,
			)
	}
	if err := validateHistoricalProductionResult(
		result,
		schedule,
	); err != nil {
		return projectionproduction.Result{},
			err
	}

	return result.Clone(), nil
}

func verifyHistoricalEndpoint(
	ctx context.Context,
	app *fiber.App,
	schedule verificationSchedule,
) (
	response.SuccessResponse[dto.ProjectionIntelligenceResponse],
	error,
) {
	if app == nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"runtime HTTP application is required",
			)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	request := httptest.NewRequest(
		http.MethodGet,
		projectionRequestURL(
			verificationFlights[0].
				TrajectoryID,
			schedule.AsOfTime,
			verificationDuration,
		),
		nil,
	).WithContext(ctx)

	timeoutMilliseconds := int(
		historicalHTTPTestTimeout /
			time.Millisecond,
	)
	httpResponse, err := app.Test(
		request,
		timeoutMilliseconds,
	)
	if err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"execute historical Projection Intelligence request within %s: %w",
				historicalHTTPTestTimeout,
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

func validateHistoricalProductionResult(
	result projectionproduction.Result,
	schedule verificationSchedule,
) error {
	if err := result.Validate(); err != nil {
		return fmt.Errorf(
			"production Projection Intelligence result is invalid: %w",
			err,
		)
	}
	if result.Strategy !=
		projectionproduction.
			StrategyHistoricalNeighbor ||
		strings.TrimSpace(
			result.FallbackReason,
		) != "" ||
		result.Projection.Method.Name !=
			projectioncontinuation.MethodName {
		return fmt.Errorf(
			"production service did not authorize Historical Neighbor Continuation: strategy=%q fallback=%q method=%q",
			result.Strategy,
			result.FallbackReason,
			result.Projection.Method.Name,
		)
	}
	if !result.Projection.Horizon.AsOfTime.Equal(
		schedule.AsOfTime,
	) ||
		result.Projection.Horizon.Duration() !=
			verificationDuration ||
		len(result.Projection.Points) != 6 {
		return fmt.Errorf(
			"production service returned an unexpected projection horizon or point count",
		)
	}
	policy := projectionread.DefaultPolicy()
	expectedNeighborCount :=
		policy.Neighbors.SelectionLimit
	if result.NeighborSelection == nil ||
		result.NeighborSelection.Status !=
			projectionneighbors.StatusComplete ||
		len(result.NeighborSelection.Neighbors) !=
			expectedNeighborCount {
		return fmt.Errorf(
			"production service selected an incomplete historical-neighbor set: status=%v count=%d want=%d",
			result.NeighborSelection,
			neighborCount(result.NeighborSelection),
			expectedNeighborCount,
		)
	}

	neighborIDs := make(
		[]string,
		0,
		len(result.NeighborSelection.Neighbors),
	)
	for _, neighbor := range result.NeighborSelection.Neighbors {
		neighborIDs = append(
			neighborIDs,
			neighbor.TrajectoryID,
		)
	}
	if err := validateHistoricalNeighborIDValues(
		neighborIDs,
	); err != nil {
		return err
	}

	if result.PatternConfidence == nil ||
		result.PatternConfidence.Status !=
			projectionpatternconfidence.StatusComplete ||
		!result.PatternConfidence.Usable ||
		result.PatternConfidence.NeighborCount !=
			expectedNeighborCount ||
		result.Freshness == nil ||
		result.Freshness.Decision !=
			projectionfreshness.DecisionAllowed ||
		!result.Freshness.Usable ||
		result.Freshness.NeighborCount !=
			expectedNeighborCount ||
		result.Freshness.RecentNeighborCount !=
			expectedNeighborCount ||
		result.RouteFrequency == nil ||
		result.RouteFrequency.Decision !=
			projectionroutefrequency.DecisionAllowed ||
		!result.RouteFrequency.Usable ||
		result.RouteFrequency.ObservationCount <
			len(verificationFlights) ||
		result.RouteFrequency.DistinctDayCount <
			len(verificationFlights) {
		return fmt.Errorf(
			"production historical evidence is incomplete or unusable",
		)
	}
	if result.ArrivalStatus !=
		projectionproduction.ArrivalStatusAttached ||
		result.Projection.Arrival == nil ||
		result.Projection.Arrival.AirportICAOCode !=
			"ZBBB" {
		return fmt.Errorf(
			"production service did not attach Estimated Arrival to ZBBB",
		)
	}

	return nil
}

func neighborCount(
	selection *projectionneighbors.Result,
) int {
	if selection == nil {
		return 0
	}

	return len(selection.Neighbors)
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

	expectedNeighborCount :=
		projectionread.DefaultPolicy().
			Neighbors.SelectionLimit
	if data.Evidence.NeighborSelection == nil ||
		data.Evidence.NeighborSelection.Status !=
			"complete" ||
		data.Evidence.NeighborSelection.
			InputCandidateCount !=
			expectedNeighborCount ||
		data.Evidence.NeighborSelection.
			CheckedCandidateCount !=
			expectedNeighborCount ||
		data.Evidence.NeighborSelection.
			QualifiedCandidateCount !=
			expectedNeighborCount ||
		len(
			data.Evidence.NeighborSelection.
				Neighbors,
		) != expectedNeighborCount {
		return fmt.Errorf(
			"historical neighbor selection is not complete: %#v",
			data.Evidence.NeighborSelection,
		)
	}
	if err := validateSelectedHistoricalNeighborIDs(
		data.Evidence.NeighborSelection.
			Neighbors,
	); err != nil {
		return err
	}

	if data.Evidence.PatternConfidence == nil ||
		data.Evidence.PatternConfidence.Status !=
			"complete" ||
		!data.Evidence.PatternConfidence.Usable ||
		data.Evidence.PatternConfidence.
			NeighborCount != expectedNeighborCount ||
		data.Evidence.PatternConfidence.Score <= 0 {
		return fmt.Errorf(
			"pattern confidence did not authorize historical continuation: %#v",
			data.Evidence.PatternConfidence,
		)
	}
	if data.Evidence.Freshness == nil ||
		!data.Evidence.Freshness.Usable ||
		data.Evidence.Freshness.Decision !=
			"allowed" ||
		data.Evidence.Freshness.NeighborCount !=
			expectedNeighborCount ||
		data.Evidence.Freshness.
			RecentNeighborCount !=
			expectedNeighborCount {
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
			ObservationCount <
			len(verificationFlights) ||
		data.Evidence.RouteFrequency.
			DistinctDayCount <
			len(verificationFlights) {
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

func validateSelectedHistoricalNeighborIDs(
	neighbors []dto.ProjectionIntelligenceNeighbor,
) error {
	trajectoryIDs := make(
		[]string,
		0,
		len(neighbors),
	)
	for _, neighbor := range neighbors {
		trajectoryIDs = append(
			trajectoryIDs,
			neighbor.TrajectoryID,
		)
	}

	return validateHistoricalNeighborIDValues(
		trajectoryIDs,
	)
}

func validateHistoricalNeighborIDValues(
	trajectoryIDs []string,
) error {
	expected := make(
		map[string]struct{},
		len(verificationFlights)-1,
	)
	for _, flight := range verificationFlights[1:] {
		expected[flight.TrajectoryID] = struct{}{}
	}

	if len(trajectoryIDs) != len(expected) {
		return fmt.Errorf(
			"historical neighbor count %d does not match expected count %d",
			len(trajectoryIDs),
			len(expected),
		)
	}

	seen := make(
		map[string]struct{},
		len(trajectoryIDs),
	)
	for _, value := range trajectoryIDs {
		trajectoryID := strings.TrimSpace(value)
		if trajectoryID ==
			verificationFlights[0].TrajectoryID {
			return fmt.Errorf(
				"current trajectory was selected as its own historical neighbor",
			)
		}
		if _, exists := expected[trajectoryID]; !exists {
			return fmt.Errorf(
				"unexpected historical neighbor trajectory %q",
				trajectoryID,
			)
		}
		if _, exists := seen[trajectoryID]; exists {
			return fmt.Errorf(
				"duplicate historical neighbor trajectory %q",
				trajectoryID,
			)
		}
		seen[trajectoryID] = struct{}{}
	}

	if len(seen) != len(expected) {
		return fmt.Errorf(
			"historical neighbor selection omitted one or more expected trajectories",
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
