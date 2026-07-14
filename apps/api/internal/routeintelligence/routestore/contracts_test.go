package routestore

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestRecordCloneDoesNotShareResultState(
	t *testing.T,
) {
	record := Record{
		Result: routecontract.Result{
			Origin: &routecontract.EndpointInference{
				Evidence: []routecontract.Evidence{
					{
						Attributes: []routecontract.EvidenceAttribute{
							{
								Key: "original",
							},
						},
					},
				},
			},
			Limitations: []routecontract.Limitation{
				{
					Code: "original",
				},
			},
			Provenance: routecontract.Provenance{
				SourceNames: []string{
					"original",
				},
			},
		},
	}

	cloned := record.Clone()
	cloned.Result.Origin.Evidence[0].
		Attributes[0].Key = "changed"
	cloned.Result.Limitations[0].Code =
		"changed"
	cloned.Result.Provenance.SourceNames[0] =
		"changed"

	if record.Result.Origin.Evidence[0].
		Attributes[0].Key != "original" ||
		record.Result.Limitations[0].Code !=
			"original" ||
		record.Result.Provenance.
			SourceNames[0] != "original" {
		t.Fatal(
			"Record.Clone() shared mutable state",
		)
	}
}

func TestPageCloneDoesNotShareRecords(
	t *testing.T,
) {
	page := Page{
		Records: []Record{
			{
				Result: routecontract.Result{
					Limitations: []routecontract.Limitation{
						{
							Code: "original",
						},
					},
				},
			},
		},
	}

	cloned := page.Clone()
	cloned.Records[0].Result.
		Limitations[0].Code = "changed"

	if page.Records[0].Result.
		Limitations[0].Code != "original" {
		t.Fatal(
			"Page.Clone() shared record state",
		)
	}
}

func TestVersionConstantsRemainStable(
	t *testing.T,
) {
	if Version != "route-store-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if PostgresVersion !=
		"route-postgres-store-v1" {
		t.Fatalf(
			"PostgresVersion = %q",
			PostgresVersion,
		)
	}
	if PostgresExecutorVersion !=
		"route-postgres-executor-v1" {
		t.Fatalf(
			"PostgresExecutorVersion = %q",
			PostgresExecutorVersion,
		)
	}
}
