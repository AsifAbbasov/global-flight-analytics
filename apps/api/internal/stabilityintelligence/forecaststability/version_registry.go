package forecaststability

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func RegisterVersion(
	request ForecastVersionRequest,
	policy VersionPolicy,
) (RegistrationResult, error) {
	if err := policy.Validate(); err != nil {
		return RegistrationResult{}, err
	}
	normalized, err := normalizeVersionRequest(request, policy)
	if err != nil {
		return RegistrationResult{}, err
	}

	outputFingerprint := projectionOutputFingerprint(normalized.Projection)
	decisionFingerprintValue := decisionFingerprint(
		normalized.Projection,
		normalized.PolicyVersion,
		normalized.ImplementationVersion,
	)

	if normalized.Previous != nil {
		changes := versionChanges(
			*normalized.Previous,
			normalized,
			outputFingerprint,
			decisionFingerprintValue,
		)
		if len(changes) == 0 {
			result := RegistrationResult{
				SchemaVersion: SchemaVersionV1,
				Status:        ResultStatusComplete,
				Decision:      RegistrationDecisionReused,
				Record:        normalized.Previous.Clone(),
				Changes:       []VersionChange{},
				Limitations:   versionLimitations(),
				Explanations: []Explanation{{
					Code:    "identical_replay_reused",
					Message: "The projection, method, policy, implementation, input fingerprint, and output fingerprint match the previous immutable forecast version.",
				}},
				ScopeGuard:  ScopeGuardResearchOnly,
				GeneratedAt: normalized.RegisteredAt,
			}
			if err := ValidateRegistrationResult(result, policy); err != nil {
				return RegistrationResult{}, err
			}
			return result.Clone(), nil
		}

		record := newVersionRecord(
			normalized,
			normalized.Previous.Ordinal+1,
			normalized.Previous.VersionID,
			outputFingerprint,
			decisionFingerprintValue,
		)
		result := RegistrationResult{
			SchemaVersion: SchemaVersionV1,
			Status:        ResultStatusComplete,
			Decision:      RegistrationDecisionCreated,
			Record:        record,
			Changes:       changes,
			Limitations:   versionLimitations(),
			Explanations: []Explanation{{
				Code:    "successor_version_created",
				Message: "A new immutable forecast version was created because one or more decision inputs, policies, methods, implementations, horizons, or outputs changed.",
			}},
			ScopeGuard:  ScopeGuardResearchOnly,
			GeneratedAt: normalized.RegisteredAt,
		}
		if err := ValidateRegistrationResult(result, policy); err != nil {
			return RegistrationResult{}, err
		}
		return result.Clone(), nil
	}

	record := newVersionRecord(
		normalized,
		1,
		"",
		outputFingerprint,
		decisionFingerprintValue,
	)
	result := RegistrationResult{
		SchemaVersion: SchemaVersionV1,
		Status:        ResultStatusComplete,
		Decision:      RegistrationDecisionInitial,
		Record:        record,
		Changes:       []VersionChange{},
		Limitations:   versionLimitations(),
		Explanations: []Explanation{{
			Code:    "initial_version_created",
			Message: "The first immutable forecast version was created for this trajectory and projection decision chain.",
		}},
		ScopeGuard:  ScopeGuardResearchOnly,
		GeneratedAt: normalized.RegisteredAt,
	}
	if err := ValidateRegistrationResult(result, policy); err != nil {
		return RegistrationResult{}, err
	}
	return result.Clone(), nil
}

func normalizeVersionRequest(
	request ForecastVersionRequest,
	policy VersionPolicy,
) (ForecastVersionRequest, error) {
	normalized := request
	normalized.Projection = request.Projection.Clone()
	normalizeProjection(&normalized.Projection)
	normalized.PolicyVersion = strings.TrimSpace(request.PolicyVersion)
	normalized.ImplementationVersion = strings.TrimSpace(request.ImplementationVersion)
	normalized.RegisteredAt = request.RegisteredAt.UTC()
	if request.Previous != nil {
		previous := request.Previous.Clone()
		normalized.Previous = &previous
	}

	projectionReport := projectioncontract.Validate(normalized.Projection)
	if projectionReport.Status != projectioncontract.ValidationStatusValid {
		return ForecastVersionRequest{}, fmt.Errorf(
			"%w: projection contract issues=%v",
			ErrInvalidVersionRequest,
			projectionReport.Issues,
		)
	}
	if normalized.PolicyVersion == "" || normalized.ImplementationVersion == "" || normalized.RegisteredAt.IsZero() {
		return ForecastVersionRequest{}, fmt.Errorf("%w: metadata and registered-at time are required", ErrInvalidVersionRequest)
	}
	if len(normalized.PolicyVersion) > policy.MaximumMetadataLength ||
		len(normalized.ImplementationVersion) > policy.MaximumMetadataLength ||
		len(normalized.Projection.Points) > policy.MaximumProjectionPointCount {
		return ForecastVersionRequest{}, fmt.Errorf("%w: capacity exceeded", ErrInvalidVersionRequest)
	}
	if normalized.RegisteredAt.Before(normalized.Projection.GeneratedAt) {
		return ForecastVersionRequest{}, fmt.Errorf("%w: registration precedes projection generation", ErrInvalidVersionRequest)
	}
	if normalized.Previous != nil {
		if err := ValidateVersionRecord(*normalized.Previous, policy); err != nil {
			return ForecastVersionRequest{}, fmt.Errorf("%w: previous record: %v", ErrInvalidVersionRequest, err)
		}
		if policy.RequireSameTrajectory && normalized.Previous.TrajectoryID != normalized.Projection.TrajectoryID {
			return ForecastVersionRequest{}, fmt.Errorf("%w: previous trajectory mismatch", ErrInvalidVersionRequest)
		}
		if policy.RequireChronologicalParent && normalized.RegisteredAt.Before(normalized.Previous.CreatedAt) {
			return ForecastVersionRequest{}, fmt.Errorf("%w: registration precedes parent version", ErrInvalidVersionRequest)
		}
	}
	return normalized, nil
}

func newVersionRecord(
	request ForecastVersionRequest,
	ordinal int,
	parentVersionID string,
	outputFingerprint string,
	decisionFingerprintValue string,
) ForecastVersionRecord {
	record := ForecastVersionRecord{
		SchemaVersion:           SchemaVersionV1,
		Ordinal:                 ordinal,
		TrajectoryID:            request.Projection.TrajectoryID,
		ProjectionSchemaVersion: request.Projection.SchemaVersion,
		Method:                  request.Projection.Method,
		PolicyVersion:           request.PolicyVersion,
		ImplementationVersion:   request.ImplementationVersion,
		InputFingerprint:        request.Projection.Provenance.InputFingerprint,
		OutputFingerprint:       outputFingerprint,
		DecisionFingerprint:     decisionFingerprintValue,
		ParentVersionID:         parentVersionID,
		Projection:              request.Projection.Clone(),
		CreatedAt:               request.RegisteredAt,
		ScopeGuard:              ScopeGuardResearchOnly,
	}
	record.VersionID = forecastVersionID(
		record.Ordinal,
		record.TrajectoryID,
		record.ParentVersionID,
		record.DecisionFingerprint,
	)
	return record
}

func versionChanges(
	previous ForecastVersionRecord,
	current ForecastVersionRequest,
	outputFingerprint string,
	decisionFingerprintValue string,
) []VersionChange {
	changes := make([]VersionChange, 0, 7)
	appendChange := func(kind VersionChangeKind, left string, right string) {
		if left != right {
			changes = append(changes, VersionChange{Kind: kind, Previous: left, Current: right})
		}
	}
	appendChange(VersionChangeProjectionSchema, string(previous.ProjectionSchemaVersion), string(current.Projection.SchemaVersion))
	appendChange(VersionChangeMethod, methodIdentity(previous.Method), methodIdentity(current.Projection.Method))
	appendChange(VersionChangePolicy, previous.PolicyVersion, current.PolicyVersion)
	appendChange(VersionChangeImplementation, previous.ImplementationVersion, current.ImplementationVersion)
	appendChange(VersionChangeInput, previous.InputFingerprint, current.Projection.Provenance.InputFingerprint)
	appendChange(VersionChangeOutput, previous.OutputFingerprint, outputFingerprint)
	appendChange(VersionChangeHorizon, horizonIdentity(previous.Projection.Horizon), horizonIdentity(current.Projection.Horizon))
	sort.Slice(changes, func(left, right int) bool { return changes[left].Kind < changes[right].Kind })
	return changes
}

func methodIdentity(method projectioncontract.Method) string {
	return strings.TrimSpace(method.Name) + "|" + strings.TrimSpace(method.Version) + "|" + string(method.DecisionClass)
}

func horizonIdentity(horizon projectioncontract.Horizon) string {
	horizon = normalizedHorizon(horizon)
	return horizon.AsOfTime.Format("20060102T150405.000000000Z") + "|" +
		horizon.EndTime.Format("20060102T150405.000000000Z") + "|" +
		strconv.FormatInt(int64(horizon.Step), 10)
}

func versionLimitations() []Limitation {
	return []Limitation{
		{
			Code:    "version_identity_not_accuracy_evidence",
			Message: "A new or reused version records forecast change history but does not prove forecast correctness or operational suitability.",
			Scope:   "forecast_accuracy",
		},
		{
			Code:    "in_memory_foundation_without_persistence",
			Message: "This foundation defines immutable version records and deterministic identifiers; durable persistence is deferred to production composition.",
			Scope:   "storage",
		},
	}
}
