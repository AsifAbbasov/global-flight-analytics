package routepipeline

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestResultCloneDoesNotShareMutableState(
	t *testing.T,
) {
	result := Result{}
	result.CatalogReport.Exclusions = append(
		result.CatalogReport.Exclusions,
		structuredExclusion(),
	)
	result.Origin.Limitations = append(
		result.Origin.Limitations,
		routecontract.Limitation{
			Code: "origin",
		},
	)
	result.Resolution.Result.Limitations =
		append(
			result.Resolution.Result.Limitations,
			routecontract.Limitation{
				Code: "route",
			},
		)
	result.Record.Result.Provenance.SourceNames =
		append(
			result.Record.Result.Provenance.SourceNames,
			"source",
		)

	cloned := result.Clone()
	cloned.CatalogReport.Exclusions[0].Count = 99
	cloned.Origin.Limitations[0].Code =
		"changed"
	cloned.Resolution.Result.Limitations[0].Code =
		"changed"
	cloned.Record.Result.Provenance.SourceNames[0] =
		"changed"

	if result.CatalogReport.Exclusions[0].Count != 1 ||
		result.Origin.Limitations[0].Code !=
			"origin" ||
		result.Resolution.Result.Limitations[0].Code !=
			"route" ||
		result.Record.Result.Provenance.SourceNames[0] !=
			"source" {
		t.Fatal(
			"Result.Clone() shared mutable state",
		)
	}
}

func TestCurrentVersionsRemainStable(
	t *testing.T,
) {
	versions := CurrentVersions()

	if versions.Pipeline !=
		"route-intelligence-pipeline-v1" ||
		versions.AirportCatalog !=
			"airport-candidate-catalog-v1" ||
		versions.AirportResolver !=
			"airport-candidate-resolver-v1" ||
		versions.EndpointEvidence !=
			"route-endpoint-evidence-v1" ||
		versions.RouteResolver !=
			"route-resolver-v1" ||
		versions.RouteStore !=
			"route-store-v1" {
		t.Fatalf(
			"unexpected versions: %#v",
			versions,
		)
	}
}

func structuredExclusion() airportresolver.ExclusionSummary {
	return airportresolver.ExclusionSummary{
		Reason: airportresolver.ExclusionReasonMissingName,
		Count:  1,
	}
}
