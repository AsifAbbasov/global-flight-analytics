package main

import "testing"

func TestCurrentNonRuntimePoliciesAreExplicit(
	t *testing.T,
) {
	expected := []string{
		modulePath + "/internal/analytics/formulabenchmark",
		modulePath + "/internal/analytics/researchbenchmark",
		modulePath + "/internal/analytics/researchdataset",
		modulePath + "/internal/analytics/transponderalert",
		modulePath + "/internal/projectionintelligence/projectionevaluation",
	}

	if len(nonRuntimePackagePolicies) != len(expected) {
		t.Fatalf(
			"policy count = %d, want %d",
			len(nonRuntimePackagePolicies),
			len(expected),
		)
	}

	for _, importPath := range expected {
		policy, exists := nonRuntimePackagePolicyFor(
			importPath,
		)
		if !exists {
			t.Fatalf(
				"missing policy for %s",
				importPath,
			)
		}
		if policy.Disposition == "" ||
			policy.Rationale == "" ||
			policy.NextAction == "" {
			t.Fatalf(
				"incomplete policy for %s: %#v",
				importPath,
				policy,
			)
		}
	}
}

func TestRemovedPackagesAreNotAllowlisted(
	t *testing.T,
) {
	removed := []string{
		modulePath + "/internal/analytics/query",
		modulePath + "/internal/analytics/window",
		modulePath + "/internal/features/datasetprofiler",
	}

	for _, importPath := range removed {
		if _, exists := nonRuntimePackagePolicyFor(
			importPath,
		); exists {
			t.Fatalf(
				"removed package remains allowlisted: %s",
				importPath,
			)
		}
	}
}

func TestUnknownNonRuntimePackageIsRejected(
	t *testing.T,
) {
	if _, exists := nonRuntimePackagePolicyFor(
		modulePath + "/internal/analytics/unknown",
	); exists {
		t.Fatal(
			"unknown package must not receive an implicit policy",
		)
	}
}

func TestProductionAirportIntelligencePackagesAreNotAllowlisted(
	t *testing.T,
) {
	productionPackages := []string{
		modulePath + "/internal/airportintelligence/history",
		modulePath + "/internal/airportintelligence/overview",
		modulePath + "/internal/airportintelligence/passport",
		modulePath + "/internal/airportintelligence/ranking",
		modulePath + "/internal/airportintelligence/statistics",
		modulePath + "/internal/airportintelligence/trends",
	}

	for _, importPath := range productionPackages {
		if _, exists := nonRuntimePackagePolicyFor(importPath); exists {
			t.Fatalf("production package remains allowlisted: %s", importPath)
		}
	}
}

// STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION

func TestProductionFeaturePackagesAreNotAllowlisted(
	t *testing.T,
) {
	productionPackages := []string{
		modulePath + "/internal/features/aircraftprovider",
		modulePath + "/internal/features/extractor",
		modulePath + "/internal/features/extractorcomposition",
		modulePath + "/internal/features/featurepipeline",
		modulePath + "/internal/features/featurestore",
		modulePath + "/internal/features/flightfeatures",
		modulePath + "/internal/features/geographicalbuilder",
		modulePath + "/internal/features/operationalbuilder",
		modulePath + "/internal/features/temporalbuilder",
		modulePath + "/internal/features/trajectorybuilder",
		modulePath + "/internal/features/validator",
	}

	for _, importPath := range productionPackages {
		if _, exists := nonRuntimePackagePolicyFor(
			importPath,
		); exists {
			t.Fatalf(
				"production Feature Pipeline package remains allowlisted: %s",
				importPath,
			)
		}
	}
}

// STAGE-14-4-FEATURE-MATERIALIZATION

// STAGE-14-6-FORMULA-BENCHMARK
