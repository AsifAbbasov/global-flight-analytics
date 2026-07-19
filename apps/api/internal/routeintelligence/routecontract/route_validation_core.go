package routecontract

import "sort"

const ValidationVersion = "route-intelligence-contract-validation-v1"

type ValidationIssue struct {
	Severity ValidationSeverity
	Code     string
	Field    string
	Message  string
}

type ValidationReport struct {
	Version      string
	Status       ValidationStatus
	ErrorCount   int
	WarningCount int
	Issues       []ValidationIssue
}

func (report ValidationReport) Clone() ValidationReport {
	cloned := report
	cloned.Issues = append(
		[]ValidationIssue(nil),
		report.Issues...,
	)

	return cloned
}

func Validate(
	result Result,
) ValidationReport {
	collector := validationCollector{}

	validateIdentity(result, &collector)
	validateWindow(result.Window, &collector)
	validateStatusAndEndpoints(result, &collector)
	validateEndpoint(
		result.Origin,
		EndpointRoleOrigin,
		result.Window.AsOfTime,
		"origin",
		&collector,
	)
	validateEndpoint(
		result.Destination,
		EndpointRoleDestination,
		result.Window.AsOfTime,
		"destination",
		&collector,
	)
	validateSummary(result, &collector)
	validateConfidence(
		result.Confidence,
		"confidence",
		totalEvidenceCount(result),
		&collector,
	)
	validateLimitations(
		result.Limitations,
		"limitations",
		&collector,
	)
	validateProvenance(result, &collector)

	sort.SliceStable(
		collector.issues,
		func(left int, right int) bool {
			leftIssue := collector.issues[left]
			rightIssue := collector.issues[right]

			if leftIssue.Field != rightIssue.Field {
				return leftIssue.Field <
					rightIssue.Field
			}
			if leftIssue.Code != rightIssue.Code {
				return leftIssue.Code <
					rightIssue.Code
			}
			if leftIssue.Severity !=
				rightIssue.Severity {
				return leftIssue.Severity <
					rightIssue.Severity
			}

			return leftIssue.Message <
				rightIssue.Message
		},
	)

	status := ValidationStatusValid
	if collector.errorCount > 0 {
		status = ValidationStatusInvalid
	}

	return ValidationReport{
		Version:      ValidationVersion,
		Status:       status,
		ErrorCount:   collector.errorCount,
		WarningCount: collector.warningCount,
		Issues: append(
			[]ValidationIssue(nil),
			collector.issues...,
		),
	}
}

type validationCollector struct {
	issues       []ValidationIssue
	errorCount   int
	warningCount int
}

func (collector *validationCollector) add(
	severity ValidationSeverity,
	code string,
	field string,
	message string,
) {
	collector.issues = append(
		collector.issues,
		ValidationIssue{
			Severity: severity,
			Code:     code,
			Field:    field,
			Message:  message,
		},
	)

	switch severity {
	case ValidationSeverityError:
		collector.errorCount++
	case ValidationSeverityWarning:
		collector.warningCount++
	}
}
