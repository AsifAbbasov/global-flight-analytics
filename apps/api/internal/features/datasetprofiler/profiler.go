package datasetprofiler

import (
	"context"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type Profiler struct {
	targetSchemaVersion flightfeatures.SchemaVersion
	now                 func() time.Time
}

func New(config Config) (*Profiler, error) {
	targetSchemaVersion := config.TargetSchemaVersion
	if targetSchemaVersion == "" {
		targetSchemaVersion =
			flightfeatures.SchemaVersionV1
	}
	if targetSchemaVersion !=
		flightfeatures.SchemaVersionV1 {
		return nil, &UnsupportedTargetSchemaError{
			SchemaVersion: targetSchemaVersion,
		}
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Profiler{
		targetSchemaVersion: targetSchemaVersion,
		now:                 now,
	}, nil
}

func (profiler *Profiler) Profile(
	ctx context.Context,
	request Request,
) (Profile, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Profile{}, err
	}

	accumulator := newProfileAccumulator(
		profiler.targetSchemaVersion,
	)
	for index, features := range request.Features {
		if index%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return Profile{}, err
			}
		}
		accumulator.add(features)
	}

	if err := ctx.Err(); err != nil {
		return Profile{}, err
	}

	profile := accumulator.profile()
	profile.Version = Version
	profile.SchemaVersion =
		profiler.targetSchemaVersion
	profile.GeneratedAt = profiler.now().UTC()

	return profile.Clone(), nil
}

type profileAccumulator struct {
	targetSchemaVersion flightfeatures.SchemaVersion
	totalRecordCount    int
	acceptedRecordCount int
	rejectedRecordCount int

	validation ValidationProfile

	trajectoryIDs map[string]struct{}
	identityKeys  map[string]struct{}
	flightIDs     map[string]struct{}
	aircraftIDs   map[string]struct{}
	icao24s       map[string]struct{}
	callsigns     map[string]struct{}

	snapshotFingerprints map[string]string
	duplicateSnapshots   int
	conflictingSnapshots int

	time timeAccumulator

	completenessScores numericAccumulator
	inputQualityScores numericAccumulator
	supportingPoints   numericAccumulator

	groups map[flightfeatures.FeatureGroup]*groupAccumulator

	sources     frequencyAccumulator
	limitations frequencyAccumulator
	rejections  frequencyAccumulator
}

func newProfileAccumulator(
	targetSchemaVersion flightfeatures.SchemaVersion,
) *profileAccumulator {
	return &profileAccumulator{
		targetSchemaVersion:  targetSchemaVersion,
		trajectoryIDs:        make(map[string]struct{}),
		identityKeys:         make(map[string]struct{}),
		flightIDs:            make(map[string]struct{}),
		aircraftIDs:          make(map[string]struct{}),
		icao24s:              make(map[string]struct{}),
		callsigns:            make(map[string]struct{}),
		snapshotFingerprints: make(map[string]string),
		groups:               newGroupAccumulators(),
		sources:              newFrequencyAccumulator(),
		limitations:          newFrequencyAccumulator(),
		rejections:           newFrequencyAccumulator(),
	}
}

func (accumulator *profileAccumulator) add(
	features flightfeatures.FlightFeatures,
) {
	accumulator.totalRecordCount++
	accumulator.addValidationStatus(
		features.Quality.Status,
	)

	rejectionReason := accumulator.rejectionReason(
		features,
	)
	if rejectionReason != "" {
		accumulator.rejectedRecordCount++
		accumulator.rejections.addRecordValues(
			[]string{rejectionReason},
		)
		return
	}

	accumulator.acceptedRecordCount++
	accumulator.addCardinality(features)
	accumulator.addSnapshot(features)
	accumulator.time.add(features)

	accumulator.completenessScores.add(
		features.Quality.CompletenessScore,
		validRatio(
			features.Quality.CompletenessScore,
		),
	)
	accumulator.inputQualityScores.add(
		features.Quality.InputQualityScore,
		validRatio(
			features.Quality.InputQualityScore,
		),
	)
	accumulator.supportingPoints.add(
		float64(features.Quality.SupportingPointCount),
		validNonNegativeInteger(
			features.Quality.SupportingPointCount,
		),
	)

	for _, group := range orderedGroups {
		accumulator.groups[group].add(
			groupEvidence(features, group),
		)
	}

	accumulator.sources.addRecordValues(
		features.Provenance.SourceNames,
	)

	limitationCodes := make(
		[]string,
		0,
	)
	for _, limitation := range allLimitations(features) {
		code := strings.TrimSpace(limitation.Code)
		if code != "" {
			limitationCodes = append(
				limitationCodes,
				code,
			)
		}
	}
	accumulator.limitations.addRecordValues(
		limitationCodes,
	)
}

func (accumulator *profileAccumulator) rejectionReason(
	features flightfeatures.FlightFeatures,
) string {
	if features.SchemaVersion !=
		accumulator.targetSchemaVersion {
		return "unsupported_schema_version"
	}

	switch features.Quality.Status {
	case flightfeatures.ValidationStatusValid,
		flightfeatures.ValidationStatusLimited:
		return ""
	case flightfeatures.ValidationStatusInvalid:
		return "validation_status_invalid"
	case flightfeatures.ValidationStatusUnvalidated:
		return "validation_status_unvalidated"
	default:
		return "validation_status_unknown"
	}
}

func (accumulator *profileAccumulator) addValidationStatus(
	status flightfeatures.ValidationStatus,
) {
	switch status {
	case flightfeatures.ValidationStatusValid:
		accumulator.validation.ValidCount++
	case flightfeatures.ValidationStatusLimited:
		accumulator.validation.LimitedCount++
	case flightfeatures.ValidationStatusInvalid:
		accumulator.validation.InvalidCount++
	case flightfeatures.ValidationStatusUnvalidated:
		accumulator.validation.UnvalidatedCount++
	default:
		accumulator.validation.UnknownCount++
	}
}

func (accumulator *profileAccumulator) addCardinality(
	features flightfeatures.FlightFeatures,
) {
	addNonEmpty(
		accumulator.trajectoryIDs,
		features.TrajectoryID,
	)
	addNonEmpty(
		accumulator.identityKeys,
		features.IdentityKey,
	)
	addNonEmpty(
		accumulator.flightIDs,
		features.FlightID,
	)
	addNonEmpty(
		accumulator.aircraftIDs,
		features.AircraftID,
	)
	addNonEmpty(
		accumulator.icao24s,
		features.ICAO24,
	)
	addNonEmpty(
		accumulator.callsigns,
		features.Callsign,
	)
}

func (accumulator *profileAccumulator) addSnapshot(
	features flightfeatures.FlightFeatures,
) {
	key := snapshotKey(features)
	fingerprint := strings.TrimSpace(
		features.Provenance.InputFingerprint,
	)

	existingFingerprint, exists :=
		accumulator.snapshotFingerprints[key]
	if !exists {
		accumulator.snapshotFingerprints[key] =
			fingerprint
		return
	}

	accumulator.duplicateSnapshots++
	if existingFingerprint != fingerprint {
		accumulator.conflictingSnapshots++
	}
}

func (accumulator profileAccumulator) profile() Profile {
	groups := make(
		[]GroupProfile,
		0,
		len(orderedGroups),
	)
	for _, group := range orderedGroups {
		groups = append(
			groups,
			accumulator.groups[group].profile(),
		)
	}

	return Profile{
		TotalRecordCount:         accumulator.totalRecordCount,
		AcceptedRecordCount:      accumulator.acceptedRecordCount,
		RejectedRecordCount:      accumulator.rejectedRecordCount,
		DuplicateSnapshotCount:   accumulator.duplicateSnapshots,
		ConflictingSnapshotCount: accumulator.conflictingSnapshots,
		Cardinality: CardinalityProfile{
			UniqueTrajectoryCount: len(accumulator.trajectoryIDs),
			UniqueIdentityCount:   len(accumulator.identityKeys),
			UniqueFlightCount:     len(accumulator.flightIDs),
			UniqueAircraftCount:   len(accumulator.aircraftIDs),
			UniqueICAO24Count:     len(accumulator.icao24s),
			UniqueCallsignCount:   len(accumulator.callsigns),
		},
		Validation: accumulator.validation,
		Time:       accumulator.time.profile(),
		Quality: QualityProfile{
			CompletenessScore: accumulator.completenessScores.profile(),
			InputQualityScore: accumulator.inputQualityScores.profile(),
			SupportingPoints:  accumulator.supportingPoints.profile(),
		},
		Groups:      groups,
		Sources:     accumulator.sources.profiles(),
		Limitations: accumulator.limitations.limitationProfiles(),
		Rejections:  accumulator.rejections.profiles(),
	}
}

func addNonEmpty(
	set map[string]struct{},
	value string,
) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return
	}

	set[normalized] = struct{}{}
}
