package featurestore

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestRecordCloneDoesNotShareFeatureSlices(t *testing.T) {
	record := Record{
		Features: flightfeatures.FlightFeatures{
			Quality: flightfeatures.FeatureQuality{
				Limitations: []flightfeatures.FeatureLimitation{
					{
						Code: "original",
					},
				},
			},
		},
	}

	cloned := record.Clone()
	cloned.Features.Quality.Limitations[0].Code = "changed"

	if record.Features.Quality.Limitations[0].Code !=
		"original" {
		t.Fatal("Record.Clone() shared feature slices")
	}
}

func TestPageCloneDoesNotShareRecords(t *testing.T) {
	page := Page{
		Records: []Record{
			{
				Features: flightfeatures.FlightFeatures{
					Quality: flightfeatures.FeatureQuality{
						Limitations: []flightfeatures.FeatureLimitation{
							{
								Code: "original",
							},
						},
					},
				},
			},
		},
		HasMore: true,
	}

	cloned := page.Clone()
	cloned.Records[0].Features.Quality.
		Limitations[0].Code = "changed"

	if page.Records[0].Features.Quality.
		Limitations[0].Code != "original" {
		t.Fatal("Page.Clone() shared record slices")
	}
	if !cloned.HasMore {
		t.Fatal("Page.Clone() did not preserve HasMore")
	}
}

func TestFeatureStoreContractConstantsRemainStable(t *testing.T) {
	if Version != "flight-feature-store-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if DefaultListLimit != 20 ||
		MaximumListLimit != 100 {
		t.Fatalf(
			"unexpected list limits: default=%d maximum=%d",
			DefaultListLimit,
			MaximumListLimit,
		)
	}
}
