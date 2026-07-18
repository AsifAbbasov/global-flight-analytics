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
	modulePath + "/internal/airportintelligence/history": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport history implementation has no production composition",
		NextAction:  "compose into Airport Intelligence read API or remove before release",
	},
	modulePath + "/internal/airportintelligence/overview": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport overview implementation has no production composition",
		NextAction:  "compose into Airport Intelligence read API or remove before release",
	},
	modulePath + "/internal/airportintelligence/passport": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport passport service has no production composition",
		NextAction:  "compose with PostgreSQL airport and analytical readers or remove before release",
	},
	modulePath + "/internal/airportintelligence/ranking": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport ranking implementation has no production composition",
		NextAction:  "compose into Airport Intelligence read API or remove before release",
	},
	modulePath + "/internal/airportintelligence/statistics": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport statistics implementation has no production composition",
		NextAction:  "compose into Airport Intelligence read API or remove before release",
	},
	modulePath + "/internal/airportintelligence/trends": {
		Disposition: dispositionPlannedProductionIntegration,
		Rationale:   "airport trends implementation has no production composition",
		NextAction:  "compose into Airport Intelligence read API or remove before release",
	},
	modulePath + "/internal/features/aircraftprovider": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/datasetprofiler": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/extractor": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/extractorcomposition": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/featurepipeline": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/featurestore": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/flightfeatures": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/geographicalbuilder": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/operationalbuilder": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/temporalbuilder": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/trajectorybuilder": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
	},
	modulePath + "/internal/features/validator": {
		Disposition: dispositionUnintegratedFeaturePipeline,
		Rationale:   "flight feature pipeline is exercised by verification tooling but not by an operational runtime root",
		NextAction:  "connect to an operational materialization command or remove the pipeline",
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
