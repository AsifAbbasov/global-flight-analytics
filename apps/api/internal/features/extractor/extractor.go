package extractor

import (
	"context"
	"errors"
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
	temporalBuilder                 TemporalBuilder
	geographicalBuilder             GeographicalBuilder
	operationalBuilder              OperationalBuilder
	trajectoryBuilder               TrajectoryBuilder
	aircraftFeatureProvider         AircraftFeatureProvider
	aircraftMetadataSourceName      string
	aircraftMetadataProviderVersion string
	fingerprintIdentity             string
	now                             func() time.Time
}

func New(config Config) (*Extractor, error) {
	if dependencyMissing(config.TemporalBuilder) {
		return nil, ErrTemporalBuilderRequired
	}
	if dependencyMissing(config.GeographicalBuilder) {
		return nil, ErrGeographicalBuilderRequired
	}
	if dependencyMissing(config.OperationalBuilder) {
		return nil, ErrOperationalBuilderRequired
	}
	if dependencyMissing(config.TrajectoryBuilder) {
		return nil, ErrTrajectoryBuilderRequired
	}

	aircraftFeatureProvider := config.AircraftFeatureProvider
	if dependencyMissing(aircraftFeatureProvider) {
		aircraftFeatureProvider = nil
	}
	aircraftMetadataSourceName := strings.TrimSpace(
		config.AircraftMetadataSourceName,
	)
	aircraftMetadataProviderVersion := strings.TrimSpace(
		config.AircraftMetadataProviderVersion,
	)
	if aircraftFeatureProvider != nil {
		if aircraftMetadataSourceName == "" {
			return nil, ErrAircraftMetadataSourceNameRequired
		}
		if aircraftMetadataProviderVersion == "" {
			return nil, ErrAircraftMetadataProviderVersionRequired
		}
	} else {
		aircraftMetadataSourceName = ""
		aircraftMetadataProviderVersion = ""
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	fingerprintIdentity := strings.TrimSpace(
		config.FingerprintIdentity,
	)
	if fingerprintIdentity == "" {
		fingerprintIdentity = Version
	}

	return &Extractor{
		temporalBuilder:                 config.TemporalBuilder,
		geographicalBuilder:             config.GeographicalBuilder,
		operationalBuilder:              config.OperationalBuilder,
		trajectoryBuilder:               config.TrajectoryBuilder,
		aircraftFeatureProvider:         aircraftFeatureProvider,
		aircraftMetadataSourceName:      aircraftMetadataSourceName,
		aircraftMetadataProviderVersion: aircraftMetadataProviderVersion,
		fingerprintIdentity:             fingerprintIdentity,
		now:                             now,
	}, nil
}

func (extractor *Extractor) Extract(
	ctx context.Context,
	request Request,
) (flightfeatures.FlightFeatures, error) {
	if ctx == nil {
		return flightfeatures.FlightFeatures{}, ErrContextRequired
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
		request.AsOfTime,
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, err
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	aircraftMetadataRetrievedAt := time.Time{}
	if extractor.aircraftFeatureProvider != nil {
		aircraftMetadataRetrievedAt = extractor.now().UTC()
	}

	fingerprint, err := fingerprintExtractionInput(
		request.Trajectory,
		aircraftFeatures,
		extractor.fingerprintIdentity,
		extractor.aircraftMetadataSourceName,
		extractor.aircraftMetadataProviderVersion,
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, err
	}

	extractedAt := extractor.now().UTC()
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
		ExtractedAt: extractedAt,

		Temporal:     temporalFeatures,
		Geographical: geographicalFeatures,
		Operational:  operationalFeatures,
		Trajectory:   trajectoryFeatures,
		Aircraft:     aircraftFeatures,

		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion: Version,
			InputFingerprint: fingerprint,
			TrajectoryCreatedAt: normalizedTrajectoryCreatedAt(
				request.Trajectory,
			),
			TrajectoryUpdatedAt: normalizedTrajectoryUpdatedAt(
				request.Trajectory,
			),
			AircraftMetadataSourceName:      extractor.aircraftMetadataSourceName,
			AircraftMetadataProviderVersion: extractor.aircraftMetadataProviderVersion,
			AircraftMetadataRetrievedAt:     aircraftMetadataRetrievedAt,
			SourceNames: collectSourceNames(
				request.Trajectory,
			),
		},
	}

	quality, err := buildInitialQuality(
		features,
		request.Trajectory,
	)
	if err != nil {
		return flightfeatures.FlightFeatures{}, err
	}
	features.Quality = quality

	return features.Clone(), nil
}

func (extractor *Extractor) buildAircraftFeatures(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
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
			AsOfTime: asOfTime.UTC(),
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
	}

	return validateSnapshotEvidence(item, request.AsOfTime)
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

func normalizedTrajectoryCreatedAt(
	item trajectory.FlightTrajectory,
) time.Time {
	if item.CreatedAt.IsZero() {
		return time.Time{}
	}
	return item.CreatedAt.UTC()
}

func normalizedTrajectoryUpdatedAt(
	item trajectory.FlightTrajectory,
) time.Time {
	if item.UpdatedAt.IsZero() {
		return time.Time{}
	}
	return item.UpdatedAt.UTC()
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
