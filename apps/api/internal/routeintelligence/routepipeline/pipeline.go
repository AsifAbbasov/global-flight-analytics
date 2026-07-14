package routepipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routeresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
)

type Pipeline struct {
	trajectoryReader TrajectoryReader
	airportLister    AirportLister
	store            routestore.Store

	endpointBuilder *endpointevidence.Builder
	routeResolver   *routeresolver.Resolver

	maximumCandidateDistanceKM float64
	maximumCandidates          int
	airportCatalogTTL          time.Duration
	airportSourceName          string
	now                        func() time.Time

	catalogCache catalogCache
}

func New(
	config Config,
) (*Pipeline, error) {
	if config.TrajectoryReader == nil {
		return nil, ErrTrajectoryReaderRequired
	}
	if config.AirportLister == nil {
		return nil, ErrAirportListerRequired
	}
	if config.Store == nil {
		return nil, ErrStoreRequired
	}

	airportCatalogTTL := config.AirportCatalogTTL
	if airportCatalogTTL == 0 {
		airportCatalogTTL =
			DefaultAirportCatalogTTL
	}
	if airportCatalogTTL < 0 {
		return nil, ErrInvalidAirportCatalogTTL
	}

	airportSourceName := strings.TrimSpace(
		config.AirportSourceName,
	)
	if airportSourceName == "" {
		airportSourceName =
			DefaultAirportSourceName
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	endpointBuilder, err := endpointevidence.New(
		config.EndpointEvidence,
	)
	if err != nil {
		return nil, &ConstructionError{
			Component: "endpoint_evidence",
			Err:       err,
		}
	}

	resolverConfig := config.RouteResolver
	resolverConfig.Now = now
	routeResolver, err := routeresolver.New(
		resolverConfig,
	)
	if err != nil {
		return nil, &ConstructionError{
			Component: "route_resolver",
			Err:       err,
		}
	}

	return &Pipeline{
		trajectoryReader: config.TrajectoryReader,
		airportLister:    config.AirportLister,
		store:            config.Store,

		endpointBuilder: endpointBuilder,
		routeResolver:   routeResolver,

		maximumCandidateDistanceKM: config.MaximumCandidateDistanceKM,
		maximumCandidates:          config.MaximumCandidates,
		airportCatalogTTL:          airportCatalogTTL,
		airportSourceName:          airportSourceName,
		now:                        now,
	}, nil
}

func (pipeline *Pipeline) Process(
	ctx context.Context,
	request Request,
) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	trajectoryID := strings.TrimSpace(
		request.TrajectoryID,
	)
	if trajectoryID == "" {
		return Result{}, ErrTrajectoryIDRequired
	}

	item, err := pipeline.trajectoryReader.
		GetTrajectoryByID(ctx, trajectoryID)
	if err != nil {
		return Result{}, newStageError(
			StageTrajectoryLoad,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	item = cloneTrajectory(item)
	item.ID = strings.TrimSpace(item.ID)
	if item.ID != trajectoryID {
		return Result{},
			newStageError(
				StageTrajectoryLoad,
				ErrTrajectoryIdentityMismatch,
			)
	}

	asOfTime := analyticalAsOfTime(item)
	if asOfTime.IsZero() {
		return Result{},
			newStageError(
				StageTrajectoryLoad,
				ErrNoAnalyticalAsOfTime,
			)
	}

	segments := usableSegments(item.Segments)
	originEvidence, destinationEvidence, catalogReport, err :=
		pipeline.buildEndpointEvidence(
			ctx,
			item,
			segments,
			asOfTime,
		)
	if err != nil {
		return Result{}, err
	}

	resolution, err := pipeline.routeResolver.Resolve(
		ctx,
		routeresolver.Input{
			TrajectoryID: item.ID,
			IdentityKey:  item.IdentityKey,
			FlightID:     item.FlightID,
			AircraftID:   item.AircraftID,
			ICAO24:       item.ICAO24,
			Callsign:     item.Callsign,
			Window: routecontract.RouteWindow{
				StartTime: item.StartTime.UTC(),
				EndTime:   item.EndTime.UTC(),
				AsOfTime:  asOfTime.UTC(),
			},
			TrajectoryUpdatedAt: trajectoryUpdatedAt(
				item,
				asOfTime,
			),
			Origin:      originEvidence.Clone(),
			Destination: destinationEvidence.Clone(),
			SourceNames: sourceNames(
				item,
				pipeline.airportSourceName,
			),
		},
	)
	if err != nil {
		return Result{}, newStageError(
			StageRouteResolution,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := Result{
		PipelineVersion: Version,
		TrajectoryID:    item.ID,
		CatalogReport:   catalogReport.Clone(),
		Origin:          originEvidence.Clone(),
		Destination:     destinationEvidence.Clone(),
		Resolution:      resolution.Clone(),
		Versions:        CurrentVersions(),
	}

	record, err := pipeline.store.Put(
		ctx,
		resolution.Result.Clone(),
	)
	if err != nil {
		return result.Clone(),
			newStageError(
				StageStorage,
				err,
			)
	}

	result.Record = record.Clone()

	return result.Clone(), nil
}

func (pipeline *Pipeline) buildEndpointEvidence(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	segments []trajectory.TrajectorySegment,
	asOfTime time.Time,
) (
	endpointevidence.Result,
	endpointevidence.Result,
	airportresolver.CatalogBuildReport,
	error,
) {
	if len(segments) == 0 {
		return unavailableEndpointEvidence(
				item.ID,
				routecontract.EndpointRoleOrigin,
				asOfTime,
				"trajectory_geometry_unavailable",
				"The persisted trajectory has no usable origin endpoint geometry.",
			),
			unavailableEndpointEvidence(
				item.ID,
				routecontract.EndpointRoleDestination,
				asOfTime,
				"trajectory_geometry_unavailable",
				"The persisted trajectory has no usable destination endpoint geometry.",
			),
			airportresolver.CatalogBuildReport{},
			nil
	}

	catalog, report, err := pipeline.loadCatalog(ctx)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageAirportCatalog,
				err,
			)
	}

	candidateResolver, err := airportresolver.New(
		airportresolver.Config{
			Catalog: catalog,
			MaximumDistanceKM: pipeline.
				maximumCandidateDistanceKM,
			MaximumCandidates: pipeline.maximumCandidates,
		},
	)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageAirportCatalog,
				err,
			)
	}

	firstSegment := segments[0]
	lastSegment := segments[len(segments)-1]

	originCandidates, err := candidateResolver.Resolve(
		ctx,
		airportresolver.Query{
			Role: routecontract.EndpointRoleOrigin,
			Point: airportresolver.Point{
				Latitude:  firstSegment.StartLatitude,
				Longitude: firstSegment.StartLongitude,
			},
		},
	)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageOriginCandidates,
				err,
			)
	}

	originEvidence, err := pipeline.endpointBuilder.Build(
		ctx,
		endpointevidence.Input{
			Candidates:        originCandidates.Clone(),
			ObservedAt:        firstSegment.StartTime.UTC(),
			TrajectoryQuality: item.QualityScore,
			SegmentStatus:     firstSegment.Status,
			SegmentPointCount: firstSegment.PointCount,
			CoverageGapCount:  item.CoverageGapCount,
		},
	)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageOriginEvidence,
				err,
			)
	}

	destinationCandidates, err :=
		candidateResolver.Resolve(
			ctx,
			airportresolver.Query{
				Role: routecontract.
					EndpointRoleDestination,
				Point: airportresolver.Point{
					Latitude:  lastSegment.EndLatitude,
					Longitude: lastSegment.EndLongitude,
				},
			},
		)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageDestinationCandidates,
				err,
			)
	}

	destinationEvidence, err :=
		pipeline.endpointBuilder.Build(
			ctx,
			endpointevidence.Input{
				Candidates:        destinationCandidates.Clone(),
				ObservedAt:        lastSegment.EndTime.UTC(),
				TrajectoryQuality: item.QualityScore,
				SegmentStatus:     lastSegment.Status,
				SegmentPointCount: lastSegment.PointCount,
				CoverageGapCount:  item.CoverageGapCount,
			},
		)
	if err != nil {
		return endpointevidence.Result{},
			endpointevidence.Result{},
			report.Clone(),
			newStageError(
				StageDestinationEvidence,
				err,
			)
	}

	return originEvidence.Clone(),
		destinationEvidence.Clone(),
		report.Clone(),
		nil
}

func newStageError(
	stage Stage,
	err error,
) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(
			err,
			context.DeadlineExceeded,
		) {
		return err
	}

	return &StageError{
		Stage: stage,
		Err:   fmt.Errorf("%w", err),
	}
}
