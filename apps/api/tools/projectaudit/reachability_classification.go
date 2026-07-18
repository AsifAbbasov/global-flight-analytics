package main

type nonRuntimeDisposition string

const (
	dispositionOfflineResearch              nonRuntimeDisposition = "offline_research"
	dispositionPlannedProductionIntegration nonRuntimeDisposition = "planned_production_integration"
	dispositionUnintegratedFeaturePipeline  nonRuntimeDisposition = "unintegrated_feature_pipeline"
	dispositionOfflineEvaluation            nonRuntimeDisposition = "offline_evaluation"
)

type nonRuntimePackagePolicy struct {
	Disposition nonRuntimeDisposition
	Rationale   string
	NextAction  string
}

var nonRuntimePackagePolicies = map[string]nonRuntimePackagePolicy{
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
	modulePath + "/internal/analytics/transponderalert": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "observed transponder evidence exists but is not yet exposed through a production read path",
		NextAction:  "integrate as read-only evidence or remove before release",
	},
	modulePath + "/internal/projectionintelligence/projectionevaluation": {
		Disposition: dispositionOfflineEvaluation,
		Rationale:   "evaluation consumes future truth and must remain separated from live forecast generation",
		NextAction:  "connect to an offline benchmark command before calibration claims",
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
