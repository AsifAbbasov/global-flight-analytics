package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualityintegration"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

const (
	productionQualityResultLimit = metricquery.MaximumResultLimit

	serverTrajectoryQuerySource = "server_trajectory_query"
)

type productionQualityInput struct {
	snapshot snapshot.Snapshot
	metadata metricexecution.PublicationMetadata
}

func (
	handler *AnalyticalMetricsHandler,
) loadProductionQualityInput(
	ctx *fiber.Ctx,
	unsupportedParameters ...string,
) (productionQualityInput, error) {
	rejected, rejectionErr :=
		rejectProductionSnapshotParameters(
			ctx,
			unsupportedParameters...,
		)
	if rejected {
		return productionQualityInput{},
			rejectionErr
	}

	recentRequest, err :=
		parseProductionQualityTrajectoryRequest(
			ctx,
		)
	if err != nil {
		return productionQualityInput{},
			analyticalBadRequest(
				ctx,
				err,
			)
	}

	selectedRegion, err := resolveAnalyticalRegion(
		ctx.Query("region"),
	)
	if err != nil {
		return productionQualityInput{},
			analyticalRegionError(
				ctx,
				err,
			)
	}

	items, err := handler.recentTrajectoriesForRegion(
		ctx.Context(),
		recentRequest,
		selectedRegion,
	)
	if err != nil {
		return productionQualityInput{},
			analyticalQueryError(
				ctx,
				err,
			)
	}

	window, err := recentRequest.Normalize(
		time.Now(),
	)
	if err != nil {
		return productionQualityInput{},
			analyticalBadRequest(
				ctx,
				err,
			)
	}

	metricSnapshot, err :=
		buildProductionQualitySnapshot(
			items,
			window,
		)
	if err != nil {
		return productionQualityInput{},
			analyticalQueryError(
				ctx,
				fmt.Errorf(
					"build server-owned analytical snapshot: %w",
					err,
				),
			)
	}

	metadata :=
		trajectoryPublicationMetadataForRegion(
			items,
			window.Limit,
			selectedRegion,
		)

	retrievedAt := time.Now().UTC()
	metadata.Sources = append(
		metadata.Sources,
		analyticalresult.Source{
			Name: serverTrajectoryQuerySource,
			Role: analyticalresult.
				SourceRoleDerived,
			ObservedFrom: window.
				ObservedFrom,
			ObservedTo: window.
				ObservedTo,
			RetrievedAt: retrievedAt,
		},
	)
	metadata.Limitations = append(
		metadata.Limitations,
		analyticalresult.Notice{
			Code:    "server_owned_production_snapshot",
			Message: "Coverage and freshness inputs were derived by the server from the retained trajectory query window.",
		},
	)

	if metricSnapshot.Time.IsZero() {
		metadata.Limitations = append(
			metadata.Limitations,
			analyticalresult.Notice{
				Code:    NoticeCodeNoProductionObservations,
				Message: "No usable retained trajectory observations were available in the server-owned query window.",
			},
		)
	}

	return productionQualityInput{
		snapshot: metricSnapshot,
		metadata: metadata,
	}, nil
}

const NoticeCodeNoProductionObservations = "no_trajectory_observations"

func rejectProductionSnapshotParameters(
	ctx *fiber.Ctx,
	names ...string,
) (bool, error) {
	for _, name := range names {
		if strings.TrimSpace(
			ctx.Query(name),
		) == "" {
			continue
		}

		return true, response.Error(
			ctx,
			fiber.StatusBadRequest,
			"PRODUCTION_SNAPSHOT_PARAMETER_NOT_SUPPORTED",
			fmt.Sprintf(
				"%s is server-owned for production quality metrics and must not be supplied",
				name,
			),
		)
	}

	if strings.TrimSpace(
		ctx.Query("limit"),
	) != "" {
		return true, response.Error(
			ctx,
			fiber.StatusBadRequest,
			"PRODUCTION_SNAPSHOT_LIMIT_NOT_SUPPORTED",
			"limit is fixed by the server for production quality metrics and must not be supplied",
		)
	}

	return false, nil
}

func parseProductionQualityTrajectoryRequest(
	ctx *fiber.Ctx,
) (metricquery.RecentRequest, error) {
	windowMinutes, err := parseOptionalInteger(
		ctx.Query("window_minutes"),
	)
	if err != nil {
		return metricquery.RecentRequest{},
			metricquery.ErrWindowMinutesInvalid
	}

	request := metricquery.RecentRequest{
		WindowMinutes: windowMinutes,
		Limit:         productionQualityResultLimit,
	}
	if _, err := request.Normalize(
		time.Now(),
	); err != nil {
		return metricquery.RecentRequest{}, err
	}

	return request, nil
}

func buildProductionQualitySnapshot(
	items []trajectory.FlightTrajectory,
	window metricquery.Window,
) (snapshot.Snapshot, error) {
	observationTimes :=
		productionObservationTimes(
			items,
			window,
		)

	density, err :=
		dataqualitycontract.
			EvaluateSamplingDensity(
				dataqualitycontract.
					SamplingDensityInput{
					WindowStart: window.
						ObservedFrom,
					WindowEnd: window.
						ObservedTo,
					ExpectedInterval: dataqualityintegration.
						DefaultExpectedObservationInterval,
					ObservationTimes: observationTimes,
				},
			)
	if err != nil {
		return snapshot.Snapshot{}, err
	}

	latestObservation := time.Time{}
	for _, observedAt := range observationTimes {
		if latestObservation.IsZero() ||
			observedAt.After(
				latestObservation,
			) {
			latestObservation =
				observedAt.UTC()
		}
	}

	return snapshot.Snapshot{
		Time: latestObservation,
		ObservedSamples: density.
			CoveredIntervalCount,
		ExpectedSamples: density.
			TotalIntervalCount,
	}, nil
}

func productionObservationTimes(
	items []trajectory.FlightTrajectory,
	window metricquery.Window,
) []time.Time {
	result := make(
		[]time.Time,
		0,
	)

	for _, item := range items {
		itemPointCount := 0

		for _, point := range item.Points {
			if !productionObservationInWindow(
				point.ObservedAt,
				window,
			) {
				continue
			}

			result = append(
				result,
				point.ObservedAt.UTC(),
			)
			itemPointCount++
		}

		if itemPointCount == 0 &&
			productionObservationInWindow(
				item.EndTime,
				window,
			) {
			result = append(
				result,
				item.EndTime.UTC(),
			)
		}
	}

	return result
}

func productionObservationInWindow(
	observedAt time.Time,
	window metricquery.Window,
) bool {
	if observedAt.IsZero() {
		return false
	}

	normalized := observedAt.UTC()
	return !normalized.Before(
		window.ObservedFrom,
	) && normalized.Before(
		window.ObservedTo,
	)
}
