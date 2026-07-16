package airspaceproduction

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

const (
	defaultWindow             = 5 * time.Minute
	minimumWindow             = 1 * time.Minute
	maximumWindow             = 1 * time.Hour
	maximumObservations       = 250000
	minimumQualityScore       = 0.35
	baseQualityScore          = 0.95
	unknownAltitudePenalty    = 0.15
	missingCallsignPenalty    = 0.05
	stationaryAirbornePenalty = 0.05
)

func New(config Config) (*Service, error) {
	if config.ObservationReader == nil {
		return nil, ErrObservationReaderRequired
	}
	if config.RegionResolver == nil {
		return nil, ErrRegionResolverRequired
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.DefaultWindow == 0 {
		config.DefaultWindow = defaultWindow
	}
	if config.MinimumWindow == 0 {
		config.MinimumWindow = minimumWindow
	}
	if config.MaximumWindow == 0 {
		config.MaximumWindow = maximumWindow
	}
	if config.MaximumObservations == 0 {
		config.MaximumObservations = maximumObservations
	}
	if config.ScenePolicy.Version == "" {
		config.ScenePolicy = localtrafficscene.DefaultPolicy()
	}
	if config.RadiusPolicy.Version == "" {
		config.RadiusPolicy = interactionradius.DefaultPolicy()
	}
	if config.ScannerPolicy.Version == "" {
		config.ScannerPolicy = proximityscanner.DefaultPolicy()
	}
	if config.RiskPolicy.Version == "" {
		config.RiskPolicy = separationrisk.DefaultPolicy()
	}
	if config.RegionPolicy.Version == "" {
		config.RegionPolicy = airspaceregionanalytics.DefaultPolicy()
	}

	if config.MinimumWindow <= 0 ||
		config.DefaultWindow < config.MinimumWindow ||
		config.MaximumWindow < config.DefaultWindow ||
		config.MaximumObservations < 1 {
		return nil, fmt.Errorf(
			"%w: production configuration bounds",
			ErrInvalidRequest,
		)
	}
	for _, validation := range []error{
		config.ScenePolicy.Validate(),
		config.RadiusPolicy.Validate(),
		config.ScannerPolicy.Validate(),
		config.RiskPolicy.Validate(),
		config.RegionPolicy.Validate(),
	} {
		if validation != nil {
			return nil, fmt.Errorf(
				"%w: policy validation: %v",
				ErrInvalidRequest,
				validation,
			)
		}
	}

	return &Service{
		observationReader:   config.ObservationReader,
		regionResolver:      config.RegionResolver,
		now:                 config.Now,
		defaultWindow:       config.DefaultWindow,
		minimumWindow:       config.MinimumWindow,
		maximumWindow:       config.MaximumWindow,
		maximumObservations: config.MaximumObservations,
		scenePolicy:         config.ScenePolicy,
		radiusPolicy:        config.RadiusPolicy,
		scannerPolicy:       config.ScannerPolicy,
		riskPolicy:          config.RiskPolicy,
		regionPolicy:        config.RegionPolicy,
	}, nil
}

func (service *Service) GetAirspaceRegionAnalytics(
	ctx context.Context,
	request Request,
) (airspaceregionanalytics.Result, error) {
	if service == nil ||
		service.observationReader == nil ||
		service.regionResolver == nil {
		return airspaceregionanalytics.Result{},
			ErrObservationReaderRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return airspaceregionanalytics.Result{}, err
	}

	normalized, generatedAt, err := service.normalizeRequest(request)
	if err != nil {
		return airspaceregionanalytics.Result{}, err
	}
	resolvedRegion, err := service.regionResolver.GetByCode(
		normalized.RegionCode,
	)
	if err != nil {
		return airspaceregionanalytics.Result{}, err
	}

	windowStart := normalized.AsOfTime.Add(-normalized.Window)
	queryStart := windowStart.Add(
		-service.radiusPolicy.MaximumObservationAge,
	)
	observations, err := service.observationReader.ListAirspaceObservations(
		ctx,
		ObservationQuery{
			Bounds:      resolvedRegion.Bounds,
			WindowStart: queryStart,
			WindowEnd:   normalized.AsOfTime,
			Limit:       service.maximumObservations + 1,
		},
	)
	if err != nil {
		return airspaceregionanalytics.Result{},
			fmt.Errorf(
				"%w: load flight-state evidence: %w",
				ErrProductionCompositionFailed,
				err,
			)
	}
	if len(observations) > service.maximumObservations {
		return airspaceregionanalytics.Result{},
			ErrObservationCapacityExceeded
	}

	snapshots, err := service.buildSnapshots(
		ctx,
		normalized.RegionCode,
		resolvedRegion.Bounds,
		windowStart,
		normalized.AsOfTime,
		generatedAt,
		observations,
	)
	if err != nil {
		return airspaceregionanalytics.Result{}, err
	}

	result, err := airspaceregionanalytics.Build(
		airspaceregionanalytics.Request{
			RegionCode:  normalized.RegionCode,
			WindowStart: windowStart,
			WindowEnd:   normalized.AsOfTime,
			GeneratedAt: generatedAt,
			Snapshots:   snapshots,
		},
		service.regionPolicy,
	)
	if err != nil {
		return airspaceregionanalytics.Result{},
			fmt.Errorf(
				"%w: build regional analytics: %w",
				ErrProductionCompositionFailed,
				err,
			)
	}
	return result.Clone(), nil
}

func (service *Service) normalizeRequest(
	request Request,
) (Request, time.Time, error) {
	normalized := request
	normalized.RegionCode = strings.ToLower(
		strings.TrimSpace(request.RegionCode),
	)
	normalized.AsOfTime = request.AsOfTime.UTC()
	if normalized.Window == 0 {
		normalized.Window = service.defaultWindow
	}

	generatedAt := service.now().UTC()
	if normalized.RegionCode == "" ||
		normalized.AsOfTime.IsZero() ||
		normalized.AsOfTime.After(generatedAt) ||
		normalized.Window < service.minimumWindow ||
		normalized.Window > service.maximumWindow ||
		normalized.Window%service.regionPolicy.TimeBucketDuration != 0 {
		return Request{}, time.Time{}, fmt.Errorf(
			"%w: region, as-of time, or window",
			ErrInvalidRequest,
		)
	}
	return normalized, generatedAt, nil
}

func (service *Service) buildSnapshots(
	ctx context.Context,
	regionCode string,
	bounds region.Bounds,
	windowStart time.Time,
	windowEnd time.Time,
	generatedAt time.Time,
	observations []Observation,
) ([]airspaceregionanalytics.SnapshotInput, error) {
	return service.buildSnapshotsWithBounds(
		ctx,
		regionCode,
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
		windowStart,
		windowEnd,
		generatedAt,
		observations,
	)
}

func (service *Service) buildSnapshotsWithBounds(
	ctx context.Context,
	regionCode string,
	minimumLatitude float64,
	maximumLatitude float64,
	minimumLongitude float64,
	maximumLongitude float64,
	windowStart time.Time,
	windowEnd time.Time,
	generatedAt time.Time,
	observations []Observation,
) ([]airspaceregionanalytics.SnapshotInput, error) {
	ordered := make([]Observation, 0, len(observations))
	for _, observation := range observations {
		ordered = append(ordered, observation.Clone())
	}
	sort.Slice(ordered, func(left, right int) bool {
		if !ordered[left].ObservedAt.Equal(ordered[right].ObservedAt) {
			return ordered[left].ObservedAt.Before(ordered[right].ObservedAt)
		}
		return observationIdentity(ordered[left]) <
			observationIdentity(ordered[right])
	})

	duration := service.regionPolicy.TimeBucketDuration
	firstSnapshotTime := windowStart.UTC().Truncate(duration).Add(duration)
	latest := make(map[string]Observation)
	observationIndex := 0
	snapshots := make([]airspaceregionanalytics.SnapshotInput, 0)

	for snapshotTime := firstSnapshotTime; !snapshotTime.After(windowEnd); snapshotTime = snapshotTime.Add(duration) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for observationIndex < len(ordered) &&
			!ordered[observationIndex].ObservedAt.After(snapshotTime) {
			observation := ordered[observationIndex]
			identity := observationIdentity(observation)
			if identity != "" {
				current, exists := latest[identity]
				if !exists ||
					observation.ObservedAt.After(current.ObservedAt) ||
					(observation.ObservedAt.Equal(current.ObservedAt) &&
						observation.StateID < current.StateID) {
					latest[identity] = observation
				}
			}
			observationIndex++
		}

		sceneInputs := make([]localtrafficscene.ObservationInput, 0, len(latest))
		for _, observation := range latest {
			age := snapshotTime.Sub(observation.ObservedAt)
			if age < 0 ||
				age > service.radiusPolicy.MaximumObservationAge {
				continue
			}
			sceneInputs = append(
				sceneInputs,
				toSceneObservation(observation),
			)
		}
		sort.Slice(sceneInputs, func(left, right int) bool {
			return sceneInputs[left].ID < sceneInputs[right].ID
		})
		if len(sceneInputs) > service.scenePolicy.MaximumInputObservationCount ||
			len(sceneInputs) > service.scannerPolicy.MaximumAircraftCount {
			return nil, ErrObservationCapacityExceeded
		}

		scene, err := localtrafficscene.Build(
			localtrafficscene.Request{
				RegionCode: regionCode,
				RegionBounds: localtrafficscene.Bounds{
					MinimumLatitude:  minimumLatitude,
					MaximumLatitude:  maximumLatitude,
					MinimumLongitude: minimumLongitude,
					MaximumLongitude: maximumLongitude,
				},
				AsOfTime:     snapshotTime,
				GeneratedAt:  generatedAt,
				Observations: sceneInputs,
			},
			service.scenePolicy,
			service.radiusPolicy,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: build local scene at %s: %w",
				ErrProductionCompositionFailed,
				snapshotTime.Format(time.RFC3339Nano),
				err,
			)
		}

		scan, err := proximityscanner.Scan(
			proximityscanner.Request{
				Scene:       scene,
				GeneratedAt: generatedAt,
			},
			service.scannerPolicy,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: scan proximity at %s: %w",
				ErrProductionCompositionFailed,
				snapshotTime.Format(time.RFC3339Nano),
				err,
			)
		}

		risk, err := separationrisk.Evaluate(
			separationrisk.Request{
				Scan:        scan,
				GeneratedAt: generatedAt,
			},
			service.riskPolicy,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: evaluate separation risk at %s: %w",
				ErrProductionCompositionFailed,
				snapshotTime.Format(time.RFC3339Nano),
				err,
			)
		}

		snapshots = append(
			snapshots,
			airspaceregionanalytics.SnapshotInput{
				Scene: scene,
				Scan:  scan,
				Risk:  risk,
			},
		)
	}
	return snapshots, nil
}

func observationIdentity(observation Observation) string {
	if value := strings.ToUpper(strings.TrimSpace(observation.ICAO24)); value != "" {
		return "icao24:" + value
	}
	if value := strings.TrimSpace(observation.AircraftID); value != "" {
		return "aircraft:" + value
	}
	return ""
}

func toSceneObservation(
	observation Observation,
) localtrafficscene.ObservationInput {
	return localtrafficscene.ObservationInput{
		ID:                          observationIdentity(observation),
		TrajectoryID:                strings.TrimSpace(observation.FlightID),
		FlightID:                    strings.TrimSpace(observation.FlightID),
		AircraftID:                  strings.TrimSpace(observation.AircraftID),
		ICAO24:                      strings.ToUpper(strings.TrimSpace(observation.ICAO24)),
		Callsign:                    strings.ToUpper(strings.TrimSpace(observation.Callsign)),
		Latitude:                    observation.Latitude,
		Longitude:                   observation.Longitude,
		AltitudeMeters:              cloneFloat64(observation.AltitudeMeters),
		AltitudeReference:           observation.AltitudeReference,
		VelocityMetersPerSecond:     observation.VelocityMetersPerSecond,
		HeadingDegrees:              observation.HeadingDegrees,
		VerticalRateMetersPerSecond: observation.VerticalRateMetersPerSecond,
		OnGround:                    observation.OnGround,
		ObservedAt:                  observation.ObservedAt.UTC(),
		SourceName:                  sourceName(observation.SourceName),
		QualityScore:                observationQuality(observation),
	}
}

func observationQuality(observation Observation) float64 {
	score := baseQualityScore
	if observation.AltitudeMeters == nil ||
		observation.AltitudeReference == interactiongraph.AltitudeReferenceUnknown {
		score -= unknownAltitudePenalty
	}
	if strings.TrimSpace(observation.Callsign) == "" {
		score -= missingCallsignPenalty
	}
	if !observation.OnGround &&
		math.Abs(observation.VelocityMetersPerSecond) < 1e-9 {
		score -= stationaryAirbornePenalty
	}
	if score < minimumQualityScore {
		return minimumQualityScore
	}
	if score > 1 {
		return 1
	}
	return score
}

func sourceName(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return postgresObservationSourceFallback
	}
	return normalized
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
