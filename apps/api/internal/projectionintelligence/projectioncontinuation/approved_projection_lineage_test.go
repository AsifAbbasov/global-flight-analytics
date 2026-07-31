package projectioncontinuation

import (
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func TestValidateApprovedProjectionLineageReconstructsFingerprintAndInputs(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	plan, err := config.HorizonPlanner.Build(projectionhorizon.Request{
		AsOfTime:          request.AsOfTime,
		RequestedDuration: request.RequestedDuration,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	evidence := ApprovedEvidence{Selection: selection, Pattern: pattern}
	result, err := baseline.ProjectApprovedWithPlan(request, plan, evidence)
	if err != nil {
		t.Fatalf("ProjectApprovedWithPlan() error = %v", err)
	}

	lineage, err := baseline.ValidateApprovedProjectionLineage(
		request,
		plan,
		evidence,
		result,
	)
	if err != nil {
		t.Fatalf("ValidateApprovedProjectionLineage() error = %v", err)
	}
	if lineage.PlanFingerprint != plan.Fingerprint ||
		lineage.SelectionFingerprint != selection.InputFingerprint ||
		lineage.PatternFingerprint != pattern.InputFingerprint ||
		lineage.ProjectionInputFingerprint != result.Provenance.InputFingerprint ||
		len(lineage.SelectedTrajectoryIDs) != len(selection.Neighbors) {
		t.Fatalf("unexpected approved lineage: %#v", lineage)
	}
}

func TestValidateApprovedProjectionLineageRejectsForeignFingerprint(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	plan, err := config.HorizonPlanner.Build(projectionhorizon.Request{
		AsOfTime:          request.AsOfTime,
		RequestedDuration: request.RequestedDuration,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	evidence := ApprovedEvidence{Selection: selection, Pattern: pattern}
	result, err := baseline.ProjectApprovedWithPlan(request, plan, evidence)
	if err != nil {
		t.Fatalf("ProjectApprovedWithPlan() error = %v", err)
	}
	result.Provenance.InputFingerprint = "sha256:" + strings.Repeat("0", 64)

	_, err = baseline.ValidateApprovedProjectionLineage(
		request,
		plan,
		evidence,
		result,
	)
	if !errors.Is(err, ErrApprovedProjectionLineageInvalid) {
		t.Fatalf("lineage error = %v", err)
	}
}

func TestValidateApprovedProjectionLineageRejectsForeignNeighborProvenance(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	plan, err := config.HorizonPlanner.Build(projectionhorizon.Request{
		AsOfTime:          request.AsOfTime,
		RequestedDuration: request.RequestedDuration,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	evidence := ApprovedEvidence{Selection: selection, Pattern: pattern}
	result, err := baseline.ProjectApprovedWithPlan(request, plan, evidence)
	if err != nil {
		t.Fatalf("ProjectApprovedWithPlan() error = %v", err)
	}
	for index := range result.Provenance.Inputs {
		if strings.HasPrefix(result.Provenance.Inputs[index].Name, "historical_neighbor:") {
			result.Provenance.Inputs[index].Name = "historical_neighbor:foreign-trajectory"
			break
		}
	}

	_, err = baseline.ValidateApprovedProjectionLineage(
		request,
		plan,
		evidence,
		result,
	)
	if !errors.Is(err, ErrApprovedProjectionLineageInvalid) {
		t.Fatalf("lineage error = %v", err)
	}
}
