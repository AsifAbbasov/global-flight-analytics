package routecontext

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	defaultMaximumCandidateDistanceKM = 120
	defaultAirportCacheTTL            = 30 * time.Minute
	meanEarthRadiusKM                 = 6371.0088
)

var (
	ErrTrajectoryReaderRequired = errors.New("trajectory reader is required")
	ErrAirportListerRequired    = errors.New("airport lister is required")
	ErrInvalidICAO24            = errors.New("invalid icao24")
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

type TrajectoryReader interface {
	GetLatestTrajectoryByICAO24(
		ctx context.Context,
		icao24 string,
	) (trajectory.FlightTrajectory, error)
}

type AirportLister interface {
	List(ctx context.Context) ([]airport.Airport, error)
}

type Config struct {
	TrajectoryReader           TrajectoryReader
	AirportLister              AirportLister
	MaximumCandidateDistanceKM float64
	AirportCacheTTL            time.Duration
	Now                        func() time.Time
}

type Service struct {
	trajectoryReader           TrajectoryReader
	airportLister              AirportLister
	maximumCandidateDistanceKM float64
	airportCacheTTL            time.Duration
	now                        func() time.Time

	airportCacheMutex     sync.Mutex
	airportCache          []airport.Airport
	airportCacheExpiresAt time.Time
	airportCacheLoaded    bool
}

func New(config Config) *Service {
	maximumCandidateDistanceKM := config.MaximumCandidateDistanceKM
	if maximumCandidateDistanceKM <= 0 ||
		math.IsNaN(maximumCandidateDistanceKM) ||
		math.IsInf(maximumCandidateDistanceKM, 0) {
		maximumCandidateDistanceKM = defaultMaximumCandidateDistanceKM
	}

	airportCacheTTL := config.AirportCacheTTL
	if airportCacheTTL <= 0 {
		airportCacheTTL = defaultAirportCacheTTL
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		trajectoryReader:           config.TrajectoryReader,
		airportLister:              config.AirportLister,
		maximumCandidateDistanceKM: maximumCandidateDistanceKM,
		airportCacheTTL:            airportCacheTTL,
		now:                        now,
	}
}

func (service *Service) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if service.trajectoryReader == nil {
		return Context{}, ErrTrajectoryReaderRequired
	}
	if service.airportLister == nil {
		return Context{}, ErrAirportListerRequired
	}

	normalizedICAO24 := strings.ToUpper(strings.TrimSpace(icao24))
	if !icao24Pattern.MatchString(normalizedICAO24) {
		return Context{}, ErrInvalidICAO24
	}

	item, err := service.trajectoryReader.GetLatestTrajectoryByICAO24(
		ctx,
		normalizedICAO24,
	)
	if err != nil {
		return Context{}, fmt.Errorf(
			"get route-context trajectory for %s: %w",
			normalizedICAO24,
			err,
		)
	}

	airports, err := service.listAirports(ctx)
	if err != nil {
		return Context{}, err
	}

	result := Context{
		ICAO24:       normalizedICAO24,
		TrajectoryID: item.ID,
		GeneratedAt:  service.now().UTC(),
	}

	segments := usableSegments(item.Segments)
	if len(segments) == 0 {
		result.Confidence = confidenceFromScore(
			0,
			[]Notice{
				{
					Code:    "no_usable_trajectory_segments",
					Message: "The latest trajectory has no usable segment endpoints for airport inference.",
				},
			},
		)
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    "route_context_unavailable",
				Message: "Probable origin and destination cannot be inferred without usable trajectory geometry.",
			},
		)

		return result, nil
	}

	firstSegment := segments[0]
	lastSegment := segments[len(segments)-1]

	result.Origin = service.findCandidate(
		airports,
		firstSegment.StartLatitude,
		firstSegment.StartLongitude,
		item.QualityScore,
		firstSegment,
		"origin",
	)
	result.Destination = service.findCandidate(
		airports,
		lastSegment.EndLatitude,
		lastSegment.EndLongitude,
		item.QualityScore,
		lastSegment,
		"destination",
	)

	result.Confidence = buildOverallConfidence(
		result.Origin,
		result.Destination,
	)
	result.Limitations = buildLimitations(
		item,
		result.Origin,
		result.Destination,
	)

	return result, nil
}

func (service *Service) listAirports(
	ctx context.Context,
) ([]airport.Airport, error) {
	service.airportCacheMutex.Lock()
	defer service.airportCacheMutex.Unlock()

	now := service.now().UTC()
	if service.airportCacheLoaded &&
		now.Before(service.airportCacheExpiresAt) {
		return append(
			[]airport.Airport(nil),
			service.airportCache...,
		), nil
	}

	items, err := service.airportLister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"list airports for route context: %w",
			err,
		)
	}

	service.airportCache = append(
		[]airport.Airport(nil),
		items...,
	)
	service.airportCacheExpiresAt = now.Add(
		service.airportCacheTTL,
	)
	service.airportCacheLoaded = true

	return append(
		[]airport.Airport(nil),
		service.airportCache...,
	), nil
}

func (service *Service) findCandidate(
	airports []airport.Airport,
	latitude float64,
	longitude float64,
	trajectoryQuality float64,
	segment trajectory.TrajectorySegment,
	role string,
) *AirportCandidate {
	var selected airport.Airport
	selectedDistanceKM := math.Inf(1)

	for _, candidate := range airports {
		if !usableAirport(candidate) {
			continue
		}

		distanceKM := haversineDistanceKM(
			latitude,
			longitude,
			candidate.Latitude,
			candidate.Longitude,
		)
		if distanceKM < selectedDistanceKM {
			selected = candidate
			selectedDistanceKM = distanceKM
		}
	}

	if math.IsInf(selectedDistanceKM, 1) ||
		selectedDistanceKM > service.maximumCandidateDistanceKM {
		return nil
	}

	distanceScore := clamp01(
		1 - selectedDistanceKM/service.maximumCandidateDistanceKM,
	)
	qualityScore := clamp01(trajectoryQuality)
	statusScore := segmentStatusScore(segment.Status)
	pointEvidenceScore := clamp01(float64(segment.PointCount) / 5)

	score := clamp01(
		0.5*distanceScore +
			0.3*qualityScore +
			0.15*statusScore +
			0.05*pointEvidenceScore,
	)

	reasons := []Notice{
		{
			Code: fmt.Sprintf("%s_nearest_airport_distance", role),
			Message: fmt.Sprintf(
				"The nearest airport is %.1f kilometres from the persisted trajectory endpoint.",
				selectedDistanceKM,
			),
		},
		{
			Code: "trajectory_quality_evidence",
			Message: fmt.Sprintf(
				"The persisted trajectory quality score contributes %.3f to the inference evidence.",
				qualityScore,
			),
		},
		{
			Code: "endpoint_segment_status",
			Message: fmt.Sprintf(
				"The endpoint segment status is %s.",
				segment.Status,
			),
		},
	}

	return &AirportCandidate{
		Airport:    selected,
		DistanceKM: selectedDistanceKM,
		Confidence: confidenceFromScore(score, reasons),
	}
}

func usableSegments(
	items []trajectory.TrajectorySegment,
) []trajectory.TrajectorySegment {
	result := make(
		[]trajectory.TrajectorySegment,
		0,
		len(items),
	)

	for _, item := range items {
		if item.Status == trajectory.SegmentStatusInvalid {
			continue
		}
		if !validCoordinates(
			item.StartLatitude,
			item.StartLongitude,
		) || !validCoordinates(
			item.EndLatitude,
			item.EndLongitude,
		) {
			continue
		}

		result = append(result, item)
	}

	sort.SliceStable(result, func(left, right int) bool {
		if result[left].SequenceNumber == result[right].SequenceNumber {
			return result[left].StartTime.Before(
				result[right].StartTime,
			)
		}

		return result[left].SequenceNumber <
			result[right].SequenceNumber
	})

	return result
}

func usableAirport(item airport.Airport) bool {
	return strings.TrimSpace(item.ICAOCode) != "" &&
		validCoordinates(item.Latitude, item.Longitude)
}

func validCoordinates(
	latitude float64,
	longitude float64,
) bool {
	return !math.IsNaN(latitude) &&
		!math.IsInf(latitude, 0) &&
		latitude >= -90 &&
		latitude <= 90 &&
		!math.IsNaN(longitude) &&
		!math.IsInf(longitude, 0) &&
		longitude >= -180 &&
		longitude <= 180
}

func buildOverallConfidence(
	origin *AirportCandidate,
	destination *AirportCandidate,
) Confidence {
	switch {
	case origin != nil && destination != nil:
		score := math.Min(
			origin.Confidence.Score,
			destination.Confidence.Score,
		)
		reasons := []Notice{
			{
				Code:    "both_route_endpoints_available",
				Message: "Both probable airport candidates are supported by persisted trajectory endpoints.",
			},
		}

		if strings.EqualFold(
			origin.Airport.ICAOCode,
			destination.Airport.ICAOCode,
		) {
			score = clamp01(score * 0.75)
			reasons = append(
				reasons,
				Notice{
					Code:    "same_airport_candidate",
					Message: "Origin and destination resolve to the same airport, so overall route confidence is reduced.",
				},
			)
		}

		return confidenceFromScore(score, reasons)

	case origin != nil || destination != nil:
		candidate := origin
		if candidate == nil {
			candidate = destination
		}

		return confidenceFromScore(
			clamp01(candidate.Confidence.Score*0.5),
			[]Notice{
				{
					Code:    "single_route_endpoint_available",
					Message: "Only one probable airport candidate is available, so complete route confidence is reduced.",
				},
			},
		)

	default:
		return confidenceFromScore(
			0,
			[]Notice{
				{
					Code:    "no_airport_candidates",
					Message: "No airport candidate is close enough to either persisted trajectory endpoint.",
				},
			},
		)
	}
}

func buildLimitations(
	item trajectory.FlightTrajectory,
	origin *AirportCandidate,
	destination *AirportCandidate,
) []Notice {
	limitations := []Notice{
		{
			Code:    "probable_route_only",
			Message: "Airport candidates are inferred from persisted trajectory endpoints and are not filed or operational flight-plan data.",
		},
		{
			Code:    "destination_not_planned_destination",
			Message: "The destination candidate reflects the latest persisted trajectory endpoint and may not be the aircraft's planned destination.",
		},
	}

	if origin == nil {
		limitations = append(
			limitations,
			Notice{
				Code:    "origin_candidate_unavailable",
				Message: "No airport is close enough to the first usable trajectory endpoint.",
			},
		)
	}
	if destination == nil {
		limitations = append(
			limitations,
			Notice{
				Code:    "destination_candidate_unavailable",
				Message: "No airport is close enough to the last usable trajectory endpoint.",
			},
		)
	}
	if item.CoverageGapCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_coverage_gaps",
				Message: fmt.Sprintf(
					"The trajectory contains %d coverage gap(s), which may weaken route-context inference.",
					item.CoverageGapCount,
				),
			},
		)
	}

	return limitations
}

func confidenceFromScore(
	score float64,
	reasons []Notice,
) Confidence {
	normalizedScore := clamp01(score)

	return Confidence{
		Score:   normalizedScore,
		Level:   confidenceLevel(normalizedScore),
		Reasons: append([]Notice(nil), reasons...),
	}
}

func confidenceLevel(score float64) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceLevelHigh
	case score >= 0.6:
		return ConfidenceLevelMedium
	case score > 0:
		return ConfidenceLevelLow
	default:
		return ConfidenceLevelNone
	}
}

func segmentStatusScore(
	status trajectory.SegmentStatus,
) float64 {
	switch status {
	case trajectory.SegmentStatusObserved:
		return 1
	case trajectory.SegmentStatusInterpolated:
		return 0.7
	case trajectory.SegmentStatusEstimated:
		return 0.45
	default:
		return 0
	}
}

func haversineDistanceKM(
	firstLatitude float64,
	firstLongitude float64,
	secondLatitude float64,
	secondLongitude float64,
) float64 {
	firstLatitudeRadians := degreesToRadians(firstLatitude)
	secondLatitudeRadians := degreesToRadians(secondLatitude)
	latitudeDelta := degreesToRadians(
		secondLatitude - firstLatitude,
	)
	longitudeDelta := degreesToRadians(
		secondLongitude - firstLongitude,
	)

	a := math.Sin(latitudeDelta/2)*math.Sin(latitudeDelta/2) +
		math.Cos(firstLatitudeRadians)*
			math.Cos(secondLatitudeRadians)*
			math.Sin(longitudeDelta/2)*
			math.Sin(longitudeDelta/2)
	a = clamp01(a)

	return meanEarthRadiusKM * 2 * math.Atan2(
		math.Sqrt(a),
		math.Sqrt(1-a),
	)
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func clamp01(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}

	return value
}
