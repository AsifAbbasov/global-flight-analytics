package projectioncontinuation

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const ApprovedProjectionLineageVersion = "historical-approved-projection-lineage-v1"

var (
	ErrApprovedProjectionLineageInvalid = errors.New(
		"approved historical projection lineage is invalid",
	)
	approvedLineageFingerprintPattern = regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
)

// ApprovedProjectionLineage is a validated receipt proving which authorized
// production evidence snapshot produced one Historical Continuation result.
type ApprovedProjectionLineage struct {
	Version string

	PlanFingerprint       string
	SelectionFingerprint  string
	PatternFingerprint    string
	SelectedTrajectoryIDs []string

	ProjectionInputFingerprint string
}

func (lineage ApprovedProjectionLineage) Clone() ApprovedProjectionLineage {
	cloned := lineage
	cloned.SelectedTrajectoryIDs = append(
		[]string(nil),
		lineage.SelectedTrajectoryIDs...,
	)
	return cloned
}

func (lineage ApprovedProjectionLineage) Validate() error {
	if lineage.Version != ApprovedProjectionLineageVersion ||
		!approvedLineageFingerprintPattern.MatchString(lineage.PlanFingerprint) ||
		!approvedLineageFingerprintPattern.MatchString(lineage.SelectionFingerprint) ||
		!approvedLineageFingerprintPattern.MatchString(lineage.PatternFingerprint) ||
		!approvedLineageFingerprintPattern.MatchString(lineage.ProjectionInputFingerprint) {
		return ErrApprovedProjectionLineageInvalid
	}
	if len(lineage.SelectedTrajectoryIDs) == 0 ||
		!sort.StringsAreSorted(lineage.SelectedTrajectoryIDs) {
		return ErrApprovedProjectionLineageInvalid
	}
	seen := make(map[string]struct{}, len(lineage.SelectedTrajectoryIDs))
	for _, trajectoryID := range lineage.SelectedTrajectoryIDs {
		normalized := strings.TrimSpace(trajectoryID)
		if normalized == "" || normalized != trajectoryID {
			return ErrApprovedProjectionLineageInvalid
		}
		if _, exists := seen[trajectoryID]; exists {
			return ErrApprovedProjectionLineageInvalid
		}
		seen[trajectoryID] = struct{}{}
	}
	return nil
}

// ValidateApprovedProjectionLineage independently reconstructs the canonical
// Historical Continuation fingerprint and provenance inputs from the supplied
// request, plan, approved Selection and Pattern, and this Baseline's immutable
// policy. It never trusts lineage asserted by a caller or adapter.
func (baseline *Baseline) ValidateApprovedProjectionLineage(
	request Request,
	plan projectionhorizon.Plan,
	evidence ApprovedEvidence,
	result projectioncontract.Result,
) (ApprovedProjectionLineage, error) {
	if baseline == nil {
		return ApprovedProjectionLineage{}, ErrHorizonPlannerRequired
	}
	if err := plan.Validate(); err != nil {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: %w",
			ErrApprovedProjectionLineageInvalid,
			err,
		)
	}

	resolution := validateApprovedContinuationEvidence(
		request,
		plan,
		evidence.Selection,
		evidence.Pattern,
	)
	if resolution.fallbackReason != "" {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: approved evidence rejected with %s",
			ErrApprovedProjectionLineageInvalid,
			resolution.fallbackReason,
		)
	}

	report := projectioncontract.Validate(result)
	if report.Status != projectioncontract.ValidationStatusValid ||
		result.Method.Name != MethodName ||
		result.Method.Version != Version ||
		result.Method.DecisionClass != projectioncontract.DecisionClassExperimental ||
		result.Status == projectioncontract.ResultStatusUnavailable {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: projection contract or historical method is invalid",
			ErrApprovedProjectionLineageInvalid,
		)
	}

	current := trajectorySnapshotAt(request.CurrentTrajectory, plan.AsOfTime)
	if len(current.Points) == 0 {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: current as-of endpoint is unavailable",
			ErrApprovedProjectionLineageInvalid,
		)
	}
	candidateByID := buildCandidateIndex(request.Candidates, plan.AsOfTime)
	if !selectedCandidateEvidenceMatches(resolution.selection, candidateByID) {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: selected candidate evidence does not match approved selection",
			ErrApprovedProjectionLineageInvalid,
		)
	}

	expectedFingerprint := continuationFingerprint(
		current,
		resolution.selection,
		resolution.pattern,
		plan,
		baseline.config,
	)
	if result.Provenance.InputFingerprint != expectedFingerprint {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: projection input fingerprint does not match approved evidence",
			ErrApprovedProjectionLineageInvalid,
		)
	}

	currentEndpoint := current.Points[len(current.Points)-1]
	expectedInputs := continuationInputs(
		currentEndpoint,
		resolution.selection,
		candidateByID,
	)
	if !sameInputReferences(result.Provenance.Inputs, expectedInputs) ||
		!result.Provenance.LatestInputObservedAt.UTC().Equal(currentEndpoint.ObservedAt.UTC()) {
		return ApprovedProjectionLineage{}, fmt.Errorf(
			"%w: projection provenance inputs do not match approved neighbors",
			ErrApprovedProjectionLineageInvalid,
		)
	}

	lineage := ApprovedProjectionLineage{
		Version:                    ApprovedProjectionLineageVersion,
		PlanFingerprint:            plan.Fingerprint,
		SelectionFingerprint:       resolution.selection.InputFingerprint,
		PatternFingerprint:         resolution.pattern.InputFingerprint,
		SelectedTrajectoryIDs:      approvedSelectedTrajectoryIDs(resolution.selection),
		ProjectionInputFingerprint: expectedFingerprint,
	}
	if err := lineage.Validate(); err != nil {
		return ApprovedProjectionLineage{}, err
	}
	return lineage.Clone(), nil
}

func approvedSelectedTrajectoryIDs(
	selection projectionneighbors.Result,
) []string {
	result := make([]string, 0, len(selection.Neighbors))
	for _, neighbor := range selection.Neighbors {
		result = append(result, strings.TrimSpace(neighbor.TrajectoryID))
	}
	sort.Strings(result)
	return result
}

func sameInputReferences(
	left []projectioncontract.InputReference,
	right []projectioncontract.InputReference,
) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
