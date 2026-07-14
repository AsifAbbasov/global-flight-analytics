package extractor

import (
	"context"
	"errors"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const aircraftFeatureFieldCount = 6

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

type Extractor struct {
	temporalBuilder         TemporalBuilder
	geographicalBuilder     GeographicalBuilder
	operationalBuilder      OperationalBuilder
	trajectoryBuilder       TrajectoryBuilder
	aircraftFeatureProvider AircraftFeatureProvider
	now                     func() time.Time
}

func New(config Config) (*Extractor, error) {
	if config.TemporalBuilder == nil {
		return nil, ErrTemporalBuilderRequired
	}
	if config.GeographicalBuilder == nil {
		return nil, ErrGeographicalBuilderRequired
	}
	if config.OperationalBuilder == nil {
		return nil, ErrOperationalBuilderRequired
	}
	if config.TrajectoryBuilder == nil {
		return nil, ErrTrajectoryBuilderRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Extractor{
		temporalBuilder:         config.TemporalBuilder,
		geographicalBuilder:     config.GeographicalBuilder,
		operationalBuilder:      config.OperationalBuilder,
		trajectoryBuilder:       config.TrajectoryBuilder,
		aircraftFeatureProvider: config.AircraftFeatureProvider,
		now:                     now,
	}, nil
}

func (extractor *Extractor) Extract(
	ctx context.Context,
	request Request,
) (flightfeatures.FlightFeatures, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}
	if err := validateRequest(request); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	temporalFeatures, err := extractor.temporalBuilder.Build(
		ctx,
		cloneTrajectory(request.Trajectory),
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, newGroupBuildError(
			flightfeatures.FeatureGroupTemporal,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	geographicalFeatures, err :=
		extractor.geographicalBuilder.Build(
			ctx,
			cloneTrajectory(request.Trajectory),
		)
	if err != nil {
		return flightfeatures.FlightFeatures{}, newGroupBuildError(
			flightfeatures.FeatureGroupGeographical,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	operationalFeatures, err := extractor.operationalBuilder.Build(
		ctx,
		cloneTrajectory(request.Trajectory),
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, newGroupBuildError(
			flightfeatures.FeatureGroupOperational,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	trajectoryFeatures, err := extractor.trajectoryBuilder.Build(
		ctx,
		cloneTrajectory(request.Trajectory),
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, newGroupBuildError(
			flightfeatures.FeatureGroupTrajectory,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	aircraftFeatures, err := extractor.buildAircraftFeatures(
		ctx,
		request.Trajectory,
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	fingerprint, err := fingerprintTrajectory(request.Trajectory)
	if err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	features := flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  request.Trajectory.ID,
		IdentityKey:   request.Trajectory.IdentityKey,
		FlightID:      request.Trajectory.FlightID,
		AircraftID:    request.Trajectory.AircraftID,
		ICAO24: strings.ToUpper(
			strings.TrimSpace(request.Trajectory.ICAO24),
		),
		Callsign: strings.TrimSpace(
			request.Trajectory.Callsign,
		),
		Window: flightfeatures.FeatureWindow{
			StartTime: request.Trajectory.StartTime.UTC(),
			EndTime:   request.Trajectory.EndTime.UTC(),
			AsOfTime:  request.AsOfTime.UTC(),
		},
		ExtractedAt: extractor.now().UTC(),

		Temporal:     temporalFeatures,
		Geographical: geographicalFeatures,
		Operational:  operationalFeatures,
		Trajectory:   trajectoryFeatures,
		Aircraft:     aircraftFeatures,

		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion: Version,
			InputFingerprint: fingerprint,
			TrajectoryUpdatedAt: normalizedTrajectoryUpdatedAt(
				request.Trajectory,
			),
			SourceNames: collectSourceNames(
				request.Trajectory,
			),
		},
	}

	features.Quality = buildInitialQuality(
		features,
		request.Trajectory,
	)

	return features.Clone(), nil
}

func (extractor *Extractor) buildAircraftFeatures(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.AircraftFeatures, error) {
	if extractor.aircraftFeatureProvider == nil {
		return flightfeatures.AircraftFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:          flightfeatures.AvailabilityStatusUnavailable,
				TotalFieldCount: aircraftFeatureFieldCount,
				Limitations: []flightfeatures.FeatureLimitation{
					{
						Code:    "aircraft_feature_provider_unavailable",
						Message: "Aircraft metadata enrichment is not configured for this extraction.",
					},
				},
			},
		}, nil
	}

	features, err := extractor.aircraftFeatureProvider.Provide(
		ctx,
		AircraftReference{
			AircraftID: strings.TrimSpace(item.AircraftID),
			ICAO24: strings.ToUpper(
				strings.TrimSpace(item.ICAO24),
			),
			Callsign: strings.TrimSpace(item.Callsign),
		},
	)
	if err != nil {
		return flightfeatures.AircraftFeatures{}, newGroupBuildError(
			flightfeatures.FeatureGroupAircraft,
			err,
		)
	}

	return features, nil
}

func validateRequest(request Request) error {
	item := request.Trajectory

	switch {
	case strings.TrimSpace(item.ID) == "":
		return ErrTrajectoryIDRequired
	case strings.TrimSpace(item.IdentityKey) == "":
		return ErrIdentityKeyRequired
	case !icao24Pattern.MatchString(
		strings.ToUpper(strings.TrimSpace(item.ICAO24)),
	):
		return ErrInvalidICAO24
	case item.StartTime.IsZero():
		return ErrTrajectoryStartTimeRequired
	case item.EndTime.IsZero():
		return ErrTrajectoryEndTimeRequired
	case item.EndTime.Before(item.StartTime):
		return ErrInvalidTrajectoryWindow
	case request.AsOfTime.IsZero():
		return ErrAsOfTimeRequired
	case request.AsOfTime.Before(item.EndTime):
		return ErrAsOfBeforeTrajectoryEnd
	case len(item.Points) == 0 && len(item.Segments) == 0:
		return ErrTrajectoryEvidenceRequired
	default:
		return nil
	}
}

func newGroupBuildError(
	group flightfeatures.FeatureGroup,
	err error,
) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return &GroupBuildError{
		Group: group,
		Err:   err,
	}
}

func buildInitialQuality(
	features flightfeatures.FlightFeatures,
	item trajectory.FlightTrajectory,
) flightfeatures.FeatureQuality {
	evidenceGroups := []flightfeatures.GroupEvidence{
		features.Temporal.Evidence,
		features.Geographical.Evidence,
		features.Operational.Evidence,
		features.Trajectory.Evidence,
		features.Aircraft.Evidence,
	}

	availableFieldCount := 0
	totalFieldCount := 0
	supportingPointCount := item.PointCount
	if len(item.Points) > supportingPointCount {
		supportingPointCount = len(item.Points)
	}
	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
	)
	seenLimitations := make(map[string]struct{})

	for _, evidence := range evidenceGroups {
		if evidence.AvailableFieldCount > 0 {
			availableFieldCount += evidence.AvailableFieldCount
		}
		if evidence.TotalFieldCount > 0 {
			totalFieldCount += evidence.TotalFieldCount
		}
		if evidence.SupportingPointCount > supportingPointCount {
			supportingPointCount = evidence.SupportingPointCount
		}

		for _, limitation := range evidence.Limitations {
			key := limitation.Code + "\x00" + limitation.Message
			if _, exists := seenLimitations[key]; exists {
				continue
			}
			seenLimitations[key] = struct{}{}
			limitations = append(limitations, limitation)
		}
	}

	completenessScore := 0.0
	if totalFieldCount > 0 {
		completenessScore = clamp01(
			float64(availableFieldCount) /
				float64(totalFieldCount),
		)
	}

	return flightfeatures.FeatureQuality{
		Status:               flightfeatures.ValidationStatusUnvalidated,
		CompletenessScore:    completenessScore,
		InputQualityScore:    clamp01(features.Trajectory.TrajectoryQualityScore),
		SupportingPointCount: supportingPointCount,
		Limitations:          limitations,
	}
}

func collectSourceNames(
	item trajectory.FlightTrajectory,
) []string {
	unique := make(map[string]struct{})

	addSourceName(unique, item.SourceName)

	for _, point := range item.Points {
		addSourceName(unique, point.SourceName)
	}
	for _, segment := range item.Segments {
		addSourceName(unique, segment.SourceName)
	}

	result := make([]string, 0, len(unique))
	for sourceName := range unique {
		result = append(result, sourceName)
	}
	sort.Strings(result)

	return result
}

func addSourceName(
	target map[string]struct{},
	value string,
) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return
	}

	target[normalized] = struct{}{}
}

func normalizedTrajectoryUpdatedAt(
	item trajectory.FlightTrajectory,
) time.Time {
	if !item.UpdatedAt.IsZero() {
		return item.UpdatedAt.UTC()
	}
	if !item.CreatedAt.IsZero() {
		return item.CreatedAt.UTC()
	}

	return item.EndTime.UTC()
}

func cloneTrajectory(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	cloned := item
	cloned.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	cloned.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	cloned.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	return cloned
}

func clamp01(value float64) float64 {
	switch {
	case math.IsNaN(value), math.IsInf(value, 0):
		return 0
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}
