package main

type nonRuntimeDisposition string

const (
	dispositionOfflineResearch             nonRuntimeDisposition = "offline_research"
	dispositionUnintegratedFeaturePipeline nonRuntimeDisposition = "unintegrated_feature_pipeline"
	dispositionOfflineEvaluation           nonRuntimeDisposition = "offline_evaluation"
)

type nonRuntimePackagePolicy struct {
	Disposition nonRuntimeDisposition
	Rationale   string
	NextAction  string
}

var nonRuntimePackagePolicies = map[string]nonRuntimePackagePolicy{
	modulePath + "/internal/analytics/formulabenchmark": {
		Disposition: dispositionOfflineEvaluation,
		Rationale:   "formula benchmarking consumes bounded offline future truth and must not enter production runtime",
		NextAction:  "retain behind benchmark-projection-formulas and require manual review before any formula change",
	},
	modulePath + "/internal/analytics/researchbenchmark": {
		Disposition: dispositionOfflineResearch,
		Rationale:   "bounded external dataset benchmark contracts are intentionally excluded from production runtime",
		NextAction:  "retain behind offline research boundary and execute only through bounded benchmark tooling",
	},
	modulePath + "/internal/analytics/researchdataset": {
		Disposition: dispositionOfflineResearch,
		Rationale:   "external research dataset governance must not become a production dependency",
		NextAction:  "retain behind offline research boundary",
	},
	modulePath + "/internal/projectionintelligence/projectionevaluation": {
		Disposition: dispositionOfflineEvaluation,
		Rationale:   "evaluation consumes future truth and must remain separated from live forecast generation",
		NextAction:  "retain behind benchmark-projection-formulas and require immutable benchmark evidence before manual calibration review",
	},
}

func nonRuntimePackagePolicyFor(
	importPath string,
) (
	nonRuntimePackagePolicy,
	bool,
) {
	policy, exists := nonRuntimePackagePolicies[importPath]
	return policy, exists
}

// STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION

// STAGE-14-4-FEATURE-MATERIALIZATION

// STAGE-14-6-FORMULA-BENCHMARK

// STAGE-14-10-TRANSPONDER-EVIDENCE-PRODUCTION
