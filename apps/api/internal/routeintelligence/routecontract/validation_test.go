package routecontract

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestValidateAcceptsCompleteResult(
	t *testing.T,
) {
	report := Validate(validCompleteResult())

	if report.Status != ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 0 ||
		len(report.Issues) != 0 {
		t.Fatalf(
			"Validate() report = %#v",
			report,
		)
	}
}

func TestValidateAcceptsUnavailableResult(
	t *testing.T,
) {
	report := Validate(validUnavailableResult())

	if report.Status != ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 0 {
		t.Fatalf(
			"Validate() report = %#v",
			report,
		)
	}
}

func TestValidateRejectsStatusEndpointMismatch(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Status = RouteStatusPartial

	report := Validate(result)

	assertIssue(
		t,
		report,
		"partial_route_endpoint_count_invalid",
		"status",
		ValidationSeverityError,
	)
}

func TestValidateRejectsEndpointRoleMismatch(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Origin.Role =
		EndpointRoleDestination

	report := Validate(result)

	assertIssue(
		t,
		report,
		"endpoint_role_mismatch",
		"origin.role",
		ValidationSeverityError,
	)
}

func TestValidateRejectsConfidenceMismatch(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Confidence.Level =
		ConfidenceLevelLow
	result.Origin.Confidence.EvidenceCount = 2

	report := Validate(result)

	assertIssue(
		t,
		report,
		"confidence_level_mismatch",
		"confidence.level",
		ValidationSeverityError,
	)
	assertIssue(
		t,
		report,
		"confidence_evidence_count_mismatch",
		"origin.confidence.evidence_count",
		ValidationSeverityError,
	)
}

func TestValidateRejectsFutureEvidence(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Destination.Evidence[0].ObservedAt =
		result.Window.AsOfTime.Add(time.Second)

	report := Validate(result)

	assertIssue(
		t,
		report,
		"evidence_after_as_of_time",
		"destination.evidence[0].observed_at",
		ValidationSeverityError,
	)
}

func TestValidateRejectsUnsortedEvidenceAttributes(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Origin.Evidence[0].Attributes = []EvidenceAttribute{
		{
			Key:   "role",
			Value: "origin",
		},
		{
			Key:   "distance_km",
			Value: "2.5",
		},
	}

	report := Validate(result)

	assertIssue(
		t,
		report,
		"evidence_attributes_not_sorted_unique",
		"origin.evidence[0].attributes",
		ValidationSeverityError,
	)
}

func TestValidateRejectsUnsortedProvenanceSources(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Provenance.SourceNames = []string{
		"trajectory",
		"ourairports",
	}

	report := Validate(result)

	assertIssue(
		t,
		report,
		"source_names_not_sorted_unique",
		"provenance.source_names",
		ValidationSeverityError,
	)
}

func TestValidateRejectsInvalidNumericValues(
	t *testing.T,
) {
	result := validCompleteResult()
	result.Origin.DistanceKM = math.NaN()
	result.Destination.Airport.Longitude =
		math.Inf(1)
	result.Confidence.Score = 1.1

	report := Validate(result)

	assertIssue(
		t,
		report,
		"endpoint_distance_invalid",
		"origin.distance_km",
		ValidationSeverityError,
	)
	assertIssue(
		t,
		report,
		"airport_longitude_invalid",
		"destination.airport.longitude",
		ValidationSeverityError,
	)
	assertIssue(
		t,
		report,
		"confidence_score_invalid",
		"confidence.score",
		ValidationSeverityError,
	)
}

func TestValidateWarnsAboutDistanceWithoutCompleteRoute(
	t *testing.T,
) {
	result := validUnavailableResult()
	result.Summary.GreatCircleDistanceKM = 100

	report := Validate(result)

	if report.Status != ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 1 {
		t.Fatalf(
			"Validate() report = %#v",
			report,
		)
	}
	assertIssue(
		t,
		report,
		"route_distance_without_complete_route",
		"summary.great_circle_distance_km",
		ValidationSeverityWarning,
	)
}

func TestValidateIssuesAreDeterministic(
	t *testing.T,
) {
	result := validCompleteResult()
	result.SchemaVersion = "unknown"
	result.ICAO24 = "bad"
	result.Status = "unknown"
	result.GeneratedAt = time.Time{}
	result.Provenance.SourceNames = []string{
		"z",
		"a",
	}

	first := Validate(result)
	second := Validate(result)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"validation reports differ:\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
	if first.Status != ValidationStatusInvalid ||
		first.ErrorCount == 0 {
		t.Fatalf(
			"unexpected report: %#v",
			first,
		)
	}

	for index := 1; index < len(first.Issues); index++ {
		previous := first.Issues[index-1]
		current := first.Issues[index]
		if previous.Field > current.Field {
			t.Fatalf(
				"issues are not sorted: %#v",
				first.Issues,
			)
		}
	}
}

func TestValidationReportCloneDoesNotShareIssues(
	t *testing.T,
) {
	report := Validate(validCompleteResult())
	report.Issues = []ValidationIssue{
		{
			Code: "original",
		},
	}

	cloned := report.Clone()
	cloned.Issues[0].Code = "changed"

	if report.Issues[0].Code != "original" {
		t.Fatal(
			"ValidationReport.Clone() shared issues",
		)
	}
}

func TestValidateRejectsNonUTCFields(
	t *testing.T,
) {
	result := validCompleteResult()
	fixedZone := time.FixedZone(
		"test",
		4*60*60,
	)
	result.Window.StartTime =
		result.Window.StartTime.In(fixedZone)

	report := Validate(result)

	assertIssue(
		t,
		report,
		"time_not_utc",
		"window.start_time",
		ValidationSeverityError,
	)
}

func assertIssue(
	t *testing.T,
	report ValidationReport,
	code string,
	field string,
	severity ValidationSeverity,
) {
	t.Helper()

	for _, issue := range report.Issues {
		if issue.Code == code &&
			issue.Field == field &&
			issue.Severity == severity {
			return
		}
	}

	t.Fatalf(
		"issue code=%q field=%q severity=%q not found in %#v",
		code,
		field,
		severity,
		report.Issues,
	)
}
