package confidencereport

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"

func (
	report Report,
) Clone() Report {
	result := report
	result.Factors = append(
		[]Contribution(nil),
		report.Factors...,
	)
	result.Reasons = append(
		[]analyticalresult.Notice(nil),
		report.Reasons...,
	)
	result.Warnings = append(
		[]analyticalresult.Notice(nil),
		report.Warnings...,
	)
	result.Limitations = append(
		[]analyticalresult.Notice(nil),
		report.Limitations...,
	)

	return result
}
