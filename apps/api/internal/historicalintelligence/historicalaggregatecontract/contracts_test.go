package historicalaggregatecontract

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestRecordCloneSeparatesResultSlices(
	t *testing.T,
) {
	record := Record{
		Result: historicalcontract.Result{
			Points: []historicalcontract.Point{
				{
					Value: 7,
				},
			},
			Provenance: historicalcontract.Provenance{
				SourceNames: []string{
					"source-a",
				},
			},
		},
	}

	cloned := record.Clone()
	cloned.Result.Points[0].Value = 11
	cloned.Result.Provenance.SourceNames[0] =
		"source-b"

	if record.Result.Points[0].Value != 7 {
		t.Fatal(
			"record clone mutated source points",
		)
	}
	if record.Result.Provenance.
		SourceNames[0] != "source-a" {
		t.Fatal(
			"record clone mutated source provenance",
		)
	}
}

func TestPageCloneSeparatesRecords(
	t *testing.T,
) {
	page := Page{
		Records: []Record{
			{
				StoredAt: time.Date(
					2026,
					time.July,
					19,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				Result: historicalcontract.Result{
					Points: []historicalcontract.Point{
						{
							Value: 4,
						},
					},
				},
			},
		},
		HasMore: true,
	}

	cloned := page.Clone()
	cloned.Records[0].Result.Points[0].Value = 9

	if page.Records[0].Result.Points[0].Value != 4 {
		t.Fatal(
			"page clone mutated source record",
		)
	}
	if !cloned.HasMore {
		t.Fatal(
			"page clone lost pagination evidence",
		)
	}
}

func TestSemanticErrorsRemainDistinct(
	t *testing.T,
) {
	errorsToCheck := []error{
		ErrUnsupportedSchemaVersion,
		ErrInputFingerprintRequired,
		ErrInvalidListLimit,
		ErrInvalidListCursor,
		ErrResultNotFound,
		ErrResultConflict,
		ErrScopeInvalid,
		ErrWindowRequired,
	}

	for leftIndex, left := range errorsToCheck {
		for rightIndex, right := range errorsToCheck {
			if leftIndex == rightIndex {
				continue
			}
			if errors.Is(left, right) {
				t.Fatalf(
					"semantic errors %d and %d are not distinct",
					leftIndex,
					rightIndex,
				)
			}
		}
	}
}
