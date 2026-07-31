package projectionproduction

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

var (
	ErrHistoricalProjectionSourceRequired = errors.New(
		"historical projection source is required",
	)
	ErrHistoricalProjectionLineageInvalid = errors.New(
		"historical projection output lineage is invalid",
	)
)

// HistoricalProjectionSource is the concrete Historical Continuation boundary
// consumed by the production adapter. The second method independently validates
// the projection against the source's immutable continuation policy.
type HistoricalProjectionSource interface {
	ProjectApprovedWithPlan(
		projectioncontinuation.Request,
		projectionhorizon.Plan,
		projectioncontinuation.ApprovedEvidence,
	) (projectioncontract.Result, error)
	ValidateApprovedProjectionLineage(
		projectioncontinuation.Request,
		projectionhorizon.Plan,
		projectioncontinuation.ApprovedEvidence,
		projectioncontract.Result,
	) (projectioncontinuation.ApprovedProjectionLineage, error)
}

// HistoricalProjectionOutcome keeps the position projection and its validated
// evidence receipt together. Internal kinematic fallback has no historical
// lineage and is distinguished by its method identity.
type HistoricalProjectionOutcome struct {
	Projection projectioncontract.Result
	Lineage    *projectioncontinuation.ApprovedProjectionLineage
}

func (outcome HistoricalProjectionOutcome) Clone() HistoricalProjectionOutcome {
	cloned := outcome
	cloned.Projection = outcome.Projection.Clone()
	if outcome.Lineage != nil {
		lineage := outcome.Lineage.Clone()
		cloned.Lineage = &lineage
	}
	return cloned
}

type HistoricalProjectionAdapter struct {
	source HistoricalProjectionSource
}

func (*HistoricalProjectionAdapter) historicalProjectionAdapter() {}

func NewHistoricalProjectionAdapter(
	source HistoricalProjectionSource,
) (*HistoricalProjectionAdapter, error) {
	if source == nil {
		return nil, ErrHistoricalProjectionSourceRequired
	}
	return &HistoricalProjectionAdapter{source: source}, nil
}

func (adapter *HistoricalProjectionAdapter) ProjectApprovedWithPlan(
	request projectioncontinuation.Request,
	plan projectionhorizon.Plan,
	evidence projectioncontinuation.ApprovedEvidence,
) (HistoricalProjectionOutcome, error) {
	if adapter == nil || adapter.source == nil {
		return HistoricalProjectionOutcome{}, ErrHistoricalProjectionSourceRequired
	}
	projection, err := adapter.source.ProjectApprovedWithPlan(
		request,
		plan,
		evidence,
	)
	if err != nil {
		return HistoricalProjectionOutcome{}, err
	}
	outcome := HistoricalProjectionOutcome{
		Projection: projection.Clone(),
	}
	if projection.Method.Name == projectionbaseline.MethodName {
		return outcome, nil
	}
	lineage, err := adapter.source.ValidateApprovedProjectionLineage(
		request,
		plan,
		evidence,
		projection,
	)
	if err != nil {
		return HistoricalProjectionOutcome{}, fmt.Errorf(
			"%w: %w",
			ErrHistoricalProjectionLineageInvalid,
			err,
		)
	}
	outcome.Lineage = &lineage
	return outcome.Clone(), nil
}

func (outcome HistoricalProjectionOutcome) ValidateAgainst(
	plan projectionhorizon.Plan,
	evidence projectioncontinuation.ApprovedEvidence,
) error {
	if outcome.Projection.Method.Name == projectionbaseline.MethodName {
		if outcome.Lineage != nil {
			return fmt.Errorf(
				"%w: kinematic fallback must not publish historical lineage",
				ErrHistoricalProjectionLineageInvalid,
			)
		}
		return nil
	}
	if outcome.Lineage == nil {
		return fmt.Errorf(
			"%w: historical projection requires lineage",
			ErrHistoricalProjectionLineageInvalid,
		)
	}
	lineage := outcome.Lineage.Clone()
	if err := lineage.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrHistoricalProjectionLineageInvalid, err)
	}
	expectedIDs := selectedTrajectoryIDs(evidence.Selection)
	if lineage.PlanFingerprint != plan.Fingerprint ||
		lineage.SelectionFingerprint != evidence.Selection.InputFingerprint ||
		lineage.PatternFingerprint != evidence.Pattern.InputFingerprint ||
		lineage.ProjectionInputFingerprint != outcome.Projection.Provenance.InputFingerprint ||
		!sameStrings(lineage.SelectedTrajectoryIDs, expectedIDs) {
		return fmt.Errorf(
			"%w: lineage does not match authorized plan, selection, pattern, or projection",
			ErrHistoricalProjectionLineageInvalid,
		)
	}
	if !historicalProvenanceNamesMatch(
		outcome.Projection.Provenance.Inputs,
		expectedIDs,
	) {
		return fmt.Errorf(
			"%w: projection provenance does not contain exactly the authorized neighbors",
			ErrHistoricalProjectionLineageInvalid,
		)
	}
	return nil
}

func historicalProvenanceNamesMatch(
	inputs []projectioncontract.InputReference,
	selectedIDs []string,
) bool {
	expected := []string{
		"current_trajectory_endpoint",
		"historical_neighbor_selection",
		"historical_pattern_confidence",
	}
	for _, trajectoryID := range selectedIDs {
		expected = append(expected, "historical_neighbor:"+trajectoryID)
	}
	actual := make([]string, 0, len(inputs))
	for _, input := range inputs {
		actual = append(actual, strings.TrimSpace(input.Name))
	}
	sort.Strings(actual)
	sort.Strings(expected)
	return sameStrings(actual, expected)
}
