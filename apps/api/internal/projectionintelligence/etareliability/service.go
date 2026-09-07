package etareliability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrServiceUnavailable    = errors.New("ETA reliability service is unavailable")
	ErrInvalidRequest        = errors.New("ETA reliability request is invalid")
	ErrCurrentETAUnavailable = errors.New("current ETA is unavailable")
	ErrRouteUnavailable      = errors.New("complete route evidence is unavailable")
	ErrLeadOutsidePolicy     = errors.New("current ETA lead time is outside the reliability policy")
)

type ProjectionReader interface {
	Get(context.Context, projectionread.Request) (projectionproduction.Result, error)
}

type SnapshotReader interface {
	LoadSnapshot(context.Context, projectionread.SnapshotRequest) (projectionread.Snapshot, error)
}

type ServiceConfig struct {
	ProjectionReader ProjectionReader
	SnapshotReader   SnapshotReader
	Policy           Policy
	Now              func() time.Time
}

type Service struct {
	projectionReader ProjectionReader
	snapshotReader   SnapshotReader
	policy           Policy
	now              func() time.Time
}

type sample struct {
	TrajectoryID         string
	AsOfTime             time.Time
	AbsoluteErrorSeconds float64
	IntervalCovered      bool
	ProjectionFingerprint string
}

func New(config ServiceConfig) (*Service, error) {
	if config.ProjectionReader == nil || config.SnapshotReader == nil {
		return nil, ErrServiceUnavailable
	}
	if err := config.Policy.Validate(); err != nil {
		return nil, fmt.Errorf("validate ETA reliability policy: %w", err)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		projectionReader: config.ProjectionReader,
		snapshotReader:   config.SnapshotReader,
		policy:           config.Policy,
		now:              now,
	}, nil
}

func (service *Service) Get(ctx context.Context, request projectionread.Request) (Result, error) {
	if service == nil || service.projectionReader == nil || service.snapshotReader == nil {
		return Result{}, ErrServiceUnavailable
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	trajectoryID := strings.TrimSpace(request.TrajectoryID)
	asOfTime := request.AsOfTime.UTC()
	if trajectoryID == "" || asOfTime.IsZero() || request.RequestedDuration < 0 {
		return Result{}, ErrInvalidRequest
	}

	current, err := service.projectionReader.Get(ctx, request)
	if err != nil {
		return Result{}, err
	}
	if current.Projection.Arrival == nil || current.ArrivalStatus != projectionproduction.ArrivalStatusAttached {
		return Result{}, ErrCurrentETAUnavailable
	}

	snapshot, err := service.snapshotReader.LoadSnapshot(ctx, projectionread.SnapshotRequest{
		TrajectoryID: trajectoryID,
		AsOfTime:     asOfTime,
	})
	if err != nil {
		return Result{}, err
	}
	if snapshot.Route == nil || snapshot.Route.Status != routecontract.RouteStatusComplete || snapshot.Route.Origin == nil || snapshot.Route.Destination == nil {
		return Result{}, ErrRouteUnavailable
	}

	origin := strings.ToUpper(strings.TrimSpace(snapshot.Route.Origin.Airport.ICAOCode))
	destination := strings.ToUpper(strings.TrimSpace(snapshot.Route.Destination.Airport.ICAOCode))
	currentArrival := current.Projection.Arrival
	targetLead := currentArrival.EstimatedTime.UTC().Sub(asOfTime)
	if targetLead < service.policy.MinimumLead || targetLead > service.policy.MaximumLead {
		return Result{}, ErrLeadOutsidePolicy
	}

	candidates := append([]trajectory.FlightTrajectory(nil), snapshot.HistoricalCandidates...)
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i].EndTime.UTC()
		right := candidates[j].EndTime.UTC()
		if left.Equal(right) {
			return candidates[i].ID < candidates[j].ID
		}
		return left.After(right)
	})
	if len(candidates) > service.policy.MaximumCandidateCount {
		candidates = candidates[:service.policy.MaximumCandidateCount]
	}

	samples := make([]sample, 0, len(candidates))
	rejected := 0
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		endpoint, ok := observedEndpoint(candidate)
		if !ok || distanceKM(endpoint.Latitude, endpoint.Longitude, snapshot.Route.Destination.Airport.Latitude, snapshot.Route.Destination.Airport.Longitude) > service.policy.EndpointRadiusKM {
			rejected++
			continue
		}
		historicalAsOf, ok := selectHistoricalAsOf(candidate, endpoint.ObservedAt.UTC(), targetLead, service.policy)
		if !ok {
			rejected++
			continue
		}

		historicalProjection, err := service.projectionReader.Get(ctx, projectionread.Request{
			TrajectoryID:      candidate.ID,
			AsOfTime:          historicalAsOf,
			RequestedDuration: request.RequestedDuration,
		})
		if err != nil || historicalProjection.Projection.Arrival == nil || historicalProjection.ArrivalStatus != projectionproduction.ArrivalStatusAttached {
			rejected++
			continue
		}
		if !sameMethod(current.Projection.Method, historicalProjection.Projection.Method) || strings.ToUpper(strings.TrimSpace(historicalProjection.Projection.Arrival.AirportICAOCode)) != destination {
			rejected++
			continue
		}

		absoluteErrorSeconds, intervalCovered, ok := evaluateArrivalSample(
			historicalProjection.Projection.Arrival,
			endpoint.ObservedAt.UTC(),
			destination,
		)
		if !ok {
			rejected++
			continue
		}
		samples = append(samples, sample{
			TrajectoryID:          candidate.ID,
			AsOfTime:              historicalAsOf,
			AbsoluteErrorSeconds:  absoluteErrorSeconds,
			IntervalCovered:       intervalCovered,
			ProjectionFingerprint: historicalProjection.CompositionFingerprint,
		})
	}

	status := StatusUnavailable
	var metrics *Metrics
	limitations := []Notice{{
		Code:    "trajectory_endpoint_arrival_proxy",
		Message: fmt.Sprintf("Historical arrival truth uses the last persisted trajectory observation within %.0f km of %s; it is not an official touchdown, gate or schedule timestamp.", service.policy.EndpointRadiusKM, destination),
	}}
	if len(samples) >= service.policy.MinimumSampleCount {
		status = StatusLimited
		if len(samples) >= service.policy.CompleteSampleCount {
			status = StatusComplete
		}
		value := buildMetrics(samples)
		metrics = &value
	} else {
		limitations = append(limitations, Notice{
			Code:    "insufficient_comparable_history",
			Message: fmt.Sprintf("Only %d comparable historical ETA evaluations were eligible; at least %d are required before publishing reliability metrics.", len(samples), service.policy.MinimumSampleCount),
		})
	}
	if status == StatusLimited {
		limitations = append(limitations, Notice{
			Code:    "limited_sample_size",
			Message: fmt.Sprintf("Reliability is based on %d samples; %d samples are required for complete status.", len(samples), service.policy.CompleteSampleCount),
		})
	}
	if rejected > 0 {
		limitations = append(limitations, Notice{
			Code:    "historical_candidates_rejected",
			Message: fmt.Sprintf("%d bounded historical candidates were excluded because endpoint, lead-time, route, method or arrival evidence was not comparable.", rejected),
		})
	}
	if len(snapshot.HistoricalCandidates) > len(candidates) {
		limitations = append(limitations, Notice{
			Code:    "candidate_scan_bounded",
			Message: fmt.Sprintf("The on-demand zero-cost reliability read evaluates at most %d recent route-scoped historical candidates.", service.policy.MaximumCandidateCount),
		})
	}

	generatedAt := service.now().UTC()
	result := Result{
		Version:              Version,
		Status:               status,
		TrajectoryID:         trajectoryID,
		Route:                Route{OriginICAOCode: origin, DestinationICAOCode: destination},
		Method:               current.Projection.Method,
		TargetLeadSeconds:    int64(targetLead / time.Second),
		LeadToleranceSeconds: int64(service.policy.LeadTolerance / time.Second),
		EndpointRadiusKM:     service.policy.EndpointRadiusKM,
		CandidateCount:       len(candidates),
		EligibleSampleCount:  len(samples),
		Metrics:              metrics,
		EvidenceClass:        EvidenceClass,
		Limitations:          normalizeNotices(limitations),
		GeneratedAt:          generatedAt,
	}
	result.InputFingerprint = reliabilityFingerprint(request, current, result, samples)
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("validate ETA reliability result: %w", err)
	}
	return result, nil
}

func observedEndpoint(item trajectory.FlightTrajectory) (trajectory.TrackPoint4D, bool) {
	if len(item.Points) == 0 {
		return trajectory.TrackPoint4D{}, false
	}
	var endpoint trajectory.TrackPoint4D
	found := false
	for _, point := range item.Points {
		if point.ObservedAt.IsZero() || math.IsNaN(point.Latitude) || math.IsNaN(point.Longitude) || point.Latitude < -90 || point.Latitude > 90 || point.Longitude < -180 || point.Longitude > 180 {
			continue
		}
		if !found || point.ObservedAt.After(endpoint.ObservedAt) {
			endpoint = point
			found = true
		}
	}
	return endpoint, found
}

func selectHistoricalAsOf(item trajectory.FlightTrajectory, endpointTime time.Time, targetLead time.Duration, policy Policy) (time.Time, bool) {
	points := append([]trajectory.TrackPoint4D(nil), item.Points...)
	sort.SliceStable(points, func(i, j int) bool { return points[i].ObservedAt.Before(points[j].ObservedAt) })
	if len(points) <= policy.MinimumPrefixPointCount {
		return time.Time{}, false
	}
	target := endpointTime.Add(-targetLead)
	bestDelta := time.Duration(math.MaxInt64)
	var best time.Time
	for index := policy.MinimumPrefixPointCount - 1; index < len(points)-1; index++ {
		observedAt := points[index].ObservedAt.UTC()
		if observedAt.IsZero() || !observedAt.Before(endpointTime) {
			continue
		}
		delta := observedAt.Sub(target)
		if delta < 0 {
			delta = -delta
		}
		if delta < bestDelta {
			bestDelta = delta
			best = observedAt
		}
	}
	if best.IsZero() || bestDelta > policy.LeadTolerance {
		return time.Time{}, false
	}
	return best, true
}

func evaluateArrivalSample(predicted *projectioncontract.ArrivalEstimate, actualBoundaryTime time.Time, destination string) (float64, bool, bool) {
	if predicted == nil || actualBoundaryTime.IsZero() {
		return 0, false, false
	}
	if strings.ToUpper(strings.TrimSpace(predicted.AirportICAOCode)) != strings.ToUpper(strings.TrimSpace(destination)) {
		return 0, false, false
	}
	earliest := predicted.EarliestTime.UTC()
	estimated := predicted.EstimatedTime.UTC()
	latest := predicted.LatestTime.UTC()
	actual := actualBoundaryTime.UTC()
	if earliest.IsZero() || estimated.IsZero() || latest.IsZero() || latest.Before(earliest) || estimated.Before(earliest) || estimated.After(latest) {
		return 0, false, false
	}
	absoluteErrorSeconds := math.Abs(estimated.Sub(actual).Seconds())
	intervalCovered := !actual.Before(earliest) && !actual.After(latest)
	return absoluteErrorSeconds, intervalCovered, true
}

func sameMethod(left, right projectioncontract.Method) bool {
	return left.Name == right.Name && left.Version == right.Version && left.DecisionClass == right.DecisionClass
}

func buildMetrics(samples []sample) Metrics {
	errorsSeconds := make([]float64, 0, len(samples))
	withinFive := 0
	withinTen := 0
	covered := 0
	for _, item := range samples {
		errorsSeconds = append(errorsSeconds, item.AbsoluteErrorSeconds)
		if item.AbsoluteErrorSeconds <= 5*60 {
			withinFive++
		}
		if item.AbsoluteErrorSeconds <= 10*60 {
			withinTen++
		}
		if item.IntervalCovered {
			covered++
		}
	}
	sort.Float64s(errorsSeconds)
	return Metrics{
		SampleCount:                len(samples),
		MedianAbsoluteErrorSeconds: median(errorsSeconds),
		P80AbsoluteErrorSeconds:    percentileNearestRank(errorsSeconds, 0.80),
		WithinFiveMinutesRatio:     float64(withinFive) / float64(len(samples)),
		WithinTenMinutesRatio:      float64(withinTen) / float64(len(samples)),
		IntervalCoverageRatio:      float64(covered) / float64(len(samples)),
	}
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}

func percentileNearestRank(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(math.Ceil(percentile*float64(len(values)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func distanceKM(lat1, lon1, lat2, lon2 float64) float64 {
	const radius = 6371.0088
	toRadians := func(value float64) float64 { return value * math.Pi / 180 }
	phi1, phi2 := toRadians(lat1), toRadians(lat2)
	dPhi := toRadians(lat2 - lat1)
	dLambda := toRadians(lon2 - lon1)
	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	return radius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func reliabilityFingerprint(request projectionread.Request, current projectionproduction.Result, result Result, samples []sample) string {
	digest := sha256.New()
	write := func(value string) { _, _ = digest.Write([]byte(value)); _, _ = digest.Write([]byte{0}) }
	write(Version)
	write(request.TrajectoryID)
	write(request.AsOfTime.UTC().Format(time.RFC3339Nano))
	write(fmt.Sprintf("%d", request.RequestedDuration))
	write(current.CompositionFingerprint)
	write(result.Route.OriginICAOCode)
	write(result.Route.DestinationICAOCode)
	write(result.Method.Name)
	write(result.Method.Version)
	write(string(result.Method.DecisionClass))
	write(fmt.Sprintf("%d", result.TargetLeadSeconds))
	for _, item := range samples {
		write(item.TrajectoryID)
		write(item.AsOfTime.UTC().Format(time.RFC3339Nano))
		write(fmt.Sprintf("%.6f", item.AbsoluteErrorSeconds))
		write(fmt.Sprintf("%t", item.IntervalCovered))
		write(item.ProjectionFingerprint)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

func normalizeNotices(items []Notice) []Notice {
	seen := make(map[string]Notice, len(items))
	for _, item := range items {
		code, message := strings.TrimSpace(item.Code), strings.TrimSpace(item.Message)
		if code == "" || message == "" {
			continue
		}
		seen[code+"\x00"+message] = Notice{Code: code, Message: message}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Notice, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}
