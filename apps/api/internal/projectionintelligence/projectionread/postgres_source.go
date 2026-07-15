package projectionread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (
	source *PostgresDataSource,
) LoadCurrentTrajectory(
	ctx context.Context,
	trajectoryID string,
	asOfTime time.Time,
) (trajectory.FlightTrajectory, error) {
	if source == nil ||
		source.client == nil ||
		source.trajectoryRepository == nil {
		return trajectory.FlightTrajectory{},
			ErrServiceUnavailable
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return trajectory.FlightTrajectory{},
			err
	}

	item, err :=
		source.trajectoryRepository.
			GetTrajectoryByID(
				ctx,
				strings.TrimSpace(
					trajectoryID,
				),
			)
	if errors.Is(
		err,
		trajectory.ErrNotFound,
	) ||
		errors.Is(
			err,
			pgx.ErrNoRows,
		) {
		return trajectory.FlightTrajectory{},
			ErrTrajectoryNotFound
	}
	if err != nil {
		return trajectory.FlightTrajectory{},
			fmt.Errorf(
				"read trajectory metadata: %w",
				err,
			)
	}

	asOfTime = asOfTime.UTC()
	if item.StartTime.IsZero() ||
		item.StartTime.UTC().After(asOfTime) {
		return trajectory.FlightTrajectory{},
			ErrTrajectoryNotFound
	}

	cutoff := asOfTime
	if !item.EndTime.IsZero() &&
		item.EndTime.UTC().Before(
			cutoff,
		) {
		cutoff = item.EndTime.UTC()
	}

	return source.hydrateTrajectory(
		ctx,
		item,
		cutoff,
	)
}

func (
	source *PostgresDataSource,
) LoadRoute(
	ctx context.Context,
	trajectoryID string,
	asOfTime time.Time,
) (routecontract.Result, error) {
	if source == nil ||
		source.client == nil {
		return routecontract.Result{},
			ErrServiceUnavailable
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return routecontract.Result{},
			err
	}

	var payload []byte
	err := source.client.QueryRow(
		ctx,
		routeAtOrBeforeSQL,
		strings.TrimSpace(
			trajectoryID,
		),
		string(
			routecontract.SchemaVersionV1,
		),
		asOfTime.UTC(),
	).Scan(
		&payload,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return routecontract.Result{},
			ErrRouteNotFound
	}
	if err != nil {
		return routecontract.Result{},
			fmt.Errorf(
				"query route result at or before as-of time: %w",
				err,
			)
	}

	var result routecontract.Result
	if err := json.Unmarshal(
		payload,
		&result,
	); err != nil {
		return routecontract.Result{},
			fmt.Errorf(
				"decode Route Intelligence result: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func (
	source *PostgresDataSource,
) LoadHistoricalCandidates(
	ctx context.Context,
	current trajectory.FlightTrajectory,
	route routecontract.Result,
	asOfTime time.Time,
) ([]trajectory.FlightTrajectory, error) {
	if source == nil ||
		source.client == nil ||
		source.trajectoryRepository == nil {
		return nil,
			ErrServiceUnavailable
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil,
			err
	}

	origin, destination, available :=
		completeRouteEndpoints(route)
	if !available {
		return []trajectory.FlightTrajectory{},
			nil
	}

	asOfTime = asOfTime.UTC()
	historyStart := asOfTime.Add(
		-source.policy.
			HistoricalCandidateLookback,
	)
	currentStart := current.StartTime.UTC()
	if currentStart.IsZero() ||
		currentStart.After(asOfTime) {
		currentStart = asOfTime
	}

	rows, err := source.client.Query(
		ctx,
		historicalCandidateIDsSQL,
		string(
			routecontract.SchemaVersionV1,
		),
		historyStart,
		asOfTime,
		strings.TrimSpace(current.ID),
		origin,
		destination,
		currentStart,
		source.policy.
			MaximumHistoricalCandidateCount,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"query route-scoped historical candidate identifiers: %w",
				err,
			)
	}
	defer rows.Close()

	candidateIDs := make(
		[]string,
		0,
		source.policy.
			MaximumHistoricalCandidateCount,
	)
	for rows.Next() {
		var candidateID string
		if err := rows.Scan(
			&candidateID,
		); err != nil {
			return nil,
				fmt.Errorf(
					"scan historical candidate identifier: %w",
					err,
				)
		}
		candidateIDs = append(
			candidateIDs,
			strings.TrimSpace(
				candidateID,
			),
		)
	}
	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"iterate historical candidate identifiers: %w",
				err,
			)
	}

	result := make(
		[]trajectory.FlightTrajectory,
		0,
		len(candidateIDs),
	)
	for _, candidateID := range candidateIDs {
		if candidateID == "" ||
			candidateID ==
				strings.TrimSpace(
					current.ID,
				) {
			continue
		}

		candidate, err :=
			source.trajectoryRepository.
				GetTrajectoryByID(
					ctx,
					candidateID,
				)
		if errors.Is(
			err,
			trajectory.ErrNotFound,
		) ||
			errors.Is(
				err,
				pgx.ErrNoRows,
			) {
			continue
		}
		if err != nil {
			return nil,
				fmt.Errorf(
					"read historical candidate %s: %w",
					candidateID,
					err,
				)
		}

		candidate, err =
			source.hydrateTrajectory(
				ctx,
				candidate,
				candidate.EndTime.UTC(),
			)
		if errors.Is(
			err,
			ErrTrajectoryPointLimitExceeded,
		) {
			continue
		}
		if err != nil {
			return nil,
				fmt.Errorf(
					"hydrate historical candidate %s: %w",
					candidateID,
					err,
				)
		}
		if len(candidate.Points) < 2 ||
			!candidate.EndTime.Before(
				currentStart,
			) {
			continue
		}

		result = append(
			result,
			candidate,
		)
	}

	return result, nil
}

func (
	source *PostgresDataSource,
) LoadRouteHistory(
	ctx context.Context,
	route routecontract.Result,
	asOfTime time.Time,
) (
	projectionroutefrequency.HistorySummary,
	error,
) {
	if source == nil ||
		source.client == nil {
		return projectionroutefrequency.HistorySummary{},
			ErrServiceUnavailable
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return projectionroutefrequency.HistorySummary{},
			err
	}

	origin, destination, available :=
		completeRouteEndpoints(route)
	if !available {
		return projectionroutefrequency.HistorySummary{},
			ErrRouteHistoryNotFound
	}

	asOfTime = asOfTime.UTC()
	windowStart := asOfTime.Add(
		-source.policy.RouteHistoryWindow,
	)
	recentStart := asOfTime.Add(
		-source.policy.RecentRouteWindow,
	)

	var observationCount int64
	var distinctFlightCount int64
	var distinctDayCount int64
	var recentObservationCount int64
	var lastObservedAt pgtype.Timestamptz

	err := source.client.QueryRow(
		ctx,
		routeHistorySummarySQL,
		string(
			routecontract.SchemaVersionV1,
		),
		windowStart,
		asOfTime,
		origin,
		destination,
		recentStart,
	).Scan(
		&observationCount,
		&distinctFlightCount,
		&distinctDayCount,
		&recentObservationCount,
		&lastObservedAt,
	)
	if err != nil {
		return projectionroutefrequency.HistorySummary{},
			fmt.Errorf(
				"query route-history summary: %w",
				err,
			)
	}
	if observationCount == 0 ||
		!lastObservedAt.Valid {
		return projectionroutefrequency.HistorySummary{},
			ErrRouteHistoryNotFound
	}

	counts := []int64{
		observationCount,
		distinctFlightCount,
		distinctDayCount,
		recentObservationCount,
	}
	for _, value := range counts {
		if value < 0 ||
			value > int64(math.MaxInt) {
			return projectionroutefrequency.HistorySummary{},
				fmt.Errorf(
					"route-history count is outside the supported integer range",
				)
		}
	}

	routeKey := origin + ">" +
		destination
	sourceNames := []string{
		source.policy.SourceName,
	}
	sort.Strings(sourceNames)

	summary := projectionroutefrequency.HistorySummary{
		RouteKey: routeKey,

		WindowStart: windowStart,
		WindowEnd:   asOfTime,
		AsOfTime:    asOfTime,

		ObservationCount:       int(observationCount),
		DistinctFlightCount:    int(distinctFlightCount),
		DistinctDayCount:       int(distinctDayCount),
		RecentObservationCount: int(recentObservationCount),
		LastObservedAt:         lastObservedAt.Time.UTC(),

		SourceNames: sourceNames,
	}
	summary.InputFingerprint =
		routeHistoryFingerprint(
			summary,
		)

	if err := summary.Validate(); err != nil {
		return projectionroutefrequency.HistorySummary{},
			fmt.Errorf(
				"validate route-history summary: %w",
				err,
			)
	}

	return summary.Clone(), nil
}

func (
	source *PostgresDataSource,
) hydrateTrajectory(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	cutoff time.Time,
) (trajectory.FlightTrajectory, error) {
	cutoff = cutoff.UTC()
	startTime := item.StartTime.UTC()
	if startTime.IsZero() ||
		cutoff.IsZero() ||
		cutoff.Before(startTime) {
		item.Points =
			[]trajectory.TrackPoint4D{}
		item.PointCount = 0
		item.Segments =
			[]trajectory.TrajectorySegment{}
		item.SegmentCount = 0
		item.CoverageGaps =
			[]trajectory.CoverageGap{}
		item.CoverageGapCount = 0
		item.EndTime = cutoff
		item.DurationSeconds = 0
		return item,
			nil
	}

	query := trajectoryPointsByAircraftSQL
	args := []any{
		strings.ToUpper(
			strings.TrimSpace(
				item.ICAO24,
			),
		),
		strings.TrimSpace(
			item.Callsign,
		),
		startTime,
		cutoff,
		source.policy.
			MaximumTrajectoryPointCount + 1,
	}
	if strings.TrimSpace(
		item.FlightID,
	) != "" {
		query = trajectoryPointsByFlightSQL
		args = []any{
			strings.TrimSpace(
				item.FlightID,
			),
			startTime,
			cutoff,
			source.policy.
				MaximumTrajectoryPointCount + 1,
		}
	}

	rows, err := source.client.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return trajectory.FlightTrajectory{},
			fmt.Errorf(
				"query trajectory observation points: %w",
				err,
			)
	}
	defer rows.Close()

	points := make(
		[]trajectory.TrackPoint4D,
		0,
		source.policy.
			MaximumTrajectoryPointCount,
	)
	for rows.Next() {
		point, err := scanTrackPoint(rows)
		if err != nil {
			return trajectory.FlightTrajectory{},
				fmt.Errorf(
					"scan trajectory observation point: %w",
					err,
				)
		}
		points = append(
			points,
			point,
		)
		if len(points) >
			source.policy.
				MaximumTrajectoryPointCount {
			return trajectory.FlightTrajectory{},
				ErrTrajectoryPointLimitExceeded
		}
	}
	if err := rows.Err(); err != nil {
		return trajectory.FlightTrajectory{},
			fmt.Errorf(
				"iterate trajectory observation points: %w",
				err,
			)
	}

	item.Points = points
	item.PointCount = len(points)
	item.Segments = filterSegmentsAt(
		item.Segments,
		cutoff,
	)
	item.SegmentCount = len(
		item.Segments,
	)
	item.CoverageGaps =
		filterCoverageGapsAt(
			item.CoverageGaps,
			cutoff,
		)
	item.CoverageGapCount = len(
		item.CoverageGaps,
	)

	if len(points) > 0 {
		item.StartTime =
			points[0].ObservedAt.UTC()
		item.EndTime =
			points[len(points)-1].
				ObservedAt.UTC()
		item.DurationSeconds = int64(
			item.EndTime.Sub(
				item.StartTime,
			) / time.Second,
		)
	} else {
		item.EndTime = cutoff
		if item.EndTime.Before(
			item.StartTime,
		) {
			item.StartTime =
				item.EndTime
		}
		item.DurationSeconds = int64(
			item.EndTime.Sub(
				item.StartTime,
			) / time.Second,
		)
	}

	if item.UpdatedAt.IsZero() ||
		item.UpdatedAt.After(cutoff) {
		item.UpdatedAt =
			item.EndTime
	}

	return item, nil
}

func scanTrackPoint(
	scanner rowScanner,
) (trajectory.TrackPoint4D, error) {
	var point trajectory.TrackPoint4D
	var barometricAltitude pgtype.Float8
	var geometricAltitude pgtype.Float8
	var barometricStatus string
	var geometricStatus string

	err := scanner.Scan(
		&point.ID,
		&point.FlightID,
		&point.AircraftID,
		&point.ICAO24,
		&point.Callsign,
		&point.Latitude,
		&point.Longitude,
		&barometricAltitude,
		&barometricStatus,
		&geometricAltitude,
		&geometricStatus,
		&point.VelocityMPS,
		&point.HeadingDegrees,
		&point.VerticalRateMPS,
		&point.OnGround,
		&point.OriginCountry,
		&point.ObservedAt,
		&point.SourceName,
	)
	if err != nil {
		return trajectory.TrackPoint4D{},
			err
	}

	point.FlightStateID =
		point.ID
	point.ICAO24 = strings.ToUpper(
		strings.TrimSpace(
			point.ICAO24,
		),
	)
	point.Callsign = strings.TrimSpace(
		point.Callsign,
	)
	point.ObservedAt =
		point.ObservedAt.UTC()
	point.BarometricAltitudeStatus =
		flightstate.AltitudeStatus(
			barometricStatus,
		)
	point.GeometricAltitudeStatus =
		flightstate.AltitudeStatus(
			geometricStatus,
		)
	if barometricAltitude.Valid {
		point.BarometricAltitudeM =
			barometricAltitude.Float64
	}
	if geometricAltitude.Valid {
		point.GeometricAltitudeM =
			geometricAltitude.Float64
	}

	return point, nil
}

func completeRouteEndpoints(
	route routecontract.Result,
) (string, string, bool) {
	if route.Status !=
		routecontract.RouteStatusComplete ||
		route.Origin == nil ||
		route.Destination == nil {
		return "", "", false
	}

	origin := strings.ToUpper(
		strings.TrimSpace(
			route.Origin.Airport.ICAOCode,
		),
	)
	destination := strings.ToUpper(
		strings.TrimSpace(
			route.Destination.Airport.ICAOCode,
		),
	)
	if len(origin) != 4 ||
		len(destination) != 4 {
		return "", "", false
	}

	return origin, destination, true
}

func filterSegmentsAt(
	items []trajectory.TrajectorySegment,
	cutoff time.Time,
) []trajectory.TrajectorySegment {
	result := make(
		[]trajectory.TrajectorySegment,
		0,
		len(items),
	)
	for _, item := range items {
		if item.StartTime.IsZero() ||
			item.StartTime.After(cutoff) {
			continue
		}
		if item.EndTime.After(cutoff) {
			item.EndTime = cutoff
			item.DurationSeconds = int64(
				item.EndTime.Sub(
					item.StartTime,
				) / time.Second,
			)
		}
		result = append(
			result,
			item,
		)
	}

	return result
}

func filterCoverageGapsAt(
	items []trajectory.CoverageGap,
	cutoff time.Time,
) []trajectory.CoverageGap {
	result := make(
		[]trajectory.CoverageGap,
		0,
		len(items),
	)
	for _, item := range items {
		if item.StartTime.IsZero() ||
			item.StartTime.After(cutoff) {
			continue
		}
		if item.EndTime.After(cutoff) {
			item.EndTime = cutoff
			item.DurationSeconds = int64(
				item.EndTime.Sub(
					item.StartTime,
				) / time.Second,
			)
		}
		result = append(
			result,
			item,
		)
	}

	return result
}

func routeHistoryFingerprint(
	summary projectionroutefrequency.HistorySummary,
) string {
	digest := sha256.Sum256(
		[]byte(
			fmt.Sprintf(
				"projection-route-history-summary-v1|%s|%s|%s|%s|%d|%d|%d|%d|%s|%s",
				summary.RouteKey,
				summary.WindowStart.UTC().
					Format(time.RFC3339Nano),
				summary.WindowEnd.UTC().
					Format(time.RFC3339Nano),
				summary.AsOfTime.UTC().
					Format(time.RFC3339Nano),
				summary.ObservationCount,
				summary.DistinctFlightCount,
				summary.DistinctDayCount,
				summary.RecentObservationCount,
				summary.LastObservedAt.UTC().
					Format(time.RFC3339Nano),
				strings.Join(
					summary.SourceNames,
					",",
				),
			),
		),
	)

	return "sha256:" +
		hex.EncodeToString(
			digest[:],
		)
}

func nonNilContext(
	ctx context.Context,
) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
