package airportresolver

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestNewCatalogNormalizesAndReportsExclusions(
	t *testing.T,
) {
	items := []airport.Airport{
		{
			ICAOCode:   " ubbb ",
			IATACode:   " gyd ",
			Name:       " Heydar Aliyev International Airport ",
			City:       " Baku ",
			Country:    " Azerbaijan ",
			Latitude:   40.4675,
			Longitude:  50.0467,
			ElevationM: 3,
			Timezone:   " Asia/Baku ",
		},
		{
			ICAOCode:  "bad",
			Name:      "Invalid ICAO",
			Latitude:  40,
			Longitude: 50,
		},
		{
			ICAOCode:  "UGTB",
			IATACode:  "TOOLONG",
			Name:      "Invalid IATA",
			Latitude:  41,
			Longitude: 44,
		},
		{
			ICAOCode:  "UBGZ",
			Name:      "",
			Latitude:  40,
			Longitude: 47,
		},
		{
			ICAOCode:  "UBBN",
			Name:      "Invalid coordinates",
			Latitude:  math.NaN(),
			Longitude: 45,
		},
		{
			ICAOCode:   "UBBY",
			Name:       "Invalid elevation",
			Latitude:   41,
			Longitude:  46,
			ElevationM: math.Inf(1),
		},
	}

	catalog, report, err := NewCatalog(items)
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	if catalog.Size() != 1 ||
		report.InputCount != 6 ||
		report.AcceptedCount != 1 ||
		report.ExcludedCount != 5 {
		t.Fatalf(
			"unexpected catalog/report: size=%d report=%#v",
			catalog.Size(),
			report,
		)
	}
	if !strings.HasPrefix(
		catalog.Fingerprint(),
		"sha256:",
	) || len(catalog.Fingerprint()) != 71 {
		t.Fatalf(
			"fingerprint = %q",
			catalog.Fingerprint(),
		)
	}
	if report.Fingerprint != catalog.Fingerprint() {
		t.Fatalf(
			"report fingerprint = %q, catalog = %q",
			report.Fingerprint,
			catalog.Fingerprint(),
		)
	}

	airports := catalog.Airports()
	if len(airports) != 1 {
		t.Fatalf("airports = %#v", airports)
	}
	item := airports[0]
	if item.ICAOCode != "UBBB" ||
		item.IATACode != "GYD" ||
		item.Name !=
			"Heydar Aliyev International Airport" ||
		item.City != "Baku" ||
		item.Country != "Azerbaijan" ||
		item.Timezone != "Asia/Baku" {
		t.Fatalf(
			"airport was not normalized: %#v",
			item,
		)
	}

	wantExclusions := []ExclusionSummary{
		{
			Reason: ExclusionReasonInvalidCoordinates,
			Count:  1,
		},
		{
			Reason: ExclusionReasonInvalidElevation,
			Count:  1,
		},
		{
			Reason: ExclusionReasonInvalidIATACode,
			Count:  1,
		},
		{
			Reason: ExclusionReasonInvalidICAOCode,
			Count:  1,
		},
		{
			Reason: ExclusionReasonMissingName,
			Count:  1,
		},
	}
	if !reflect.DeepEqual(
		report.Exclusions,
		wantExclusions,
	) {
		t.Fatalf(
			"exclusions = %#v, want %#v",
			report.Exclusions,
			wantExclusions,
		)
	}
}

func TestNewCatalogIsIndependentOfInputOrderAndDuplicates(
	t *testing.T,
) {
	firstInput := []airport.Airport{
		airportFixture(
			"UGTB",
			"TBS",
			"Tbilisi International Airport",
			41.6692,
			44.9547,
		),
		airportFixture(
			"UBBB",
			"GYD",
			"Zulu duplicate",
			40.4675,
			50.0467,
		),
		airportFixture(
			"UBBB",
			"GYD",
			"Alpha canonical",
			40.4675,
			50.0467,
		),
	}
	secondInput := []airport.Airport{
		firstInput[2],
		firstInput[0],
		firstInput[1],
	}

	firstCatalog, firstReport, err :=
		NewCatalog(firstInput)
	if err != nil {
		t.Fatalf(
			"first NewCatalog() error = %v",
			err,
		)
	}
	secondCatalog, secondReport, err :=
		NewCatalog(secondInput)
	if err != nil {
		t.Fatalf(
			"second NewCatalog() error = %v",
			err,
		)
	}

	if firstCatalog.Fingerprint() !=
		secondCatalog.Fingerprint() {
		t.Fatalf(
			"fingerprints differ: %q %q",
			firstCatalog.Fingerprint(),
			secondCatalog.Fingerprint(),
		)
	}
	if !reflect.DeepEqual(
		firstCatalog.Airports(),
		secondCatalog.Airports(),
	) {
		t.Fatalf(
			"catalogs differ: %#v %#v",
			firstCatalog.Airports(),
			secondCatalog.Airports(),
		)
	}
	if firstCatalog.Airports()[0].Name !=
		"Alpha canonical" {
		t.Fatalf(
			"duplicate resolution selected %q",
			firstCatalog.Airports()[0].Name,
		)
	}
	wantDuplicate := []ExclusionSummary{
		{
			Reason: ExclusionReasonDuplicateICAOCode,
			Count:  1,
		},
	}
	if !reflect.DeepEqual(
		firstReport.Exclusions,
		wantDuplicate,
	) || !reflect.DeepEqual(
		secondReport.Exclusions,
		wantDuplicate,
	) {
		t.Fatalf(
			"duplicate reports differ: %#v %#v",
			firstReport,
			secondReport,
		)
	}
}

func TestCatalogReturnsDefensiveCopies(
	t *testing.T,
) {
	catalog, report, err := NewCatalog(
		[]airport.Airport{
			airportFixture(
				"UBBB",
				"GYD",
				"Airport",
				40.4675,
				50.0467,
			),
		},
	)
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}

	airports := catalog.Airports()
	airports[0].Name = "Changed"

	next := catalog.Airports()
	if next[0].Name == "Changed" {
		t.Fatal(
			"Catalog.Airports() shared state",
		)
	}

	clonedReport := report.Clone()
	clonedReport.Exclusions = append(
		clonedReport.Exclusions,
		ExclusionSummary{
			Reason: ExclusionReasonMissingName,
			Count:  1,
		},
	)
	if len(report.Exclusions) != 0 {
		t.Fatal(
			"CatalogBuildReport.Clone() shared state",
		)
	}
}

func TestNewCatalogRejectsCatalogWithoutUsableAirports(
	t *testing.T,
) {
	catalog, report, err := NewCatalog(
		[]airport.Airport{
			{
				ICAOCode: "bad",
			},
		},
	)
	if !errors.Is(err, ErrNoUsableAirports) {
		t.Fatalf(
			"NewCatalog() error = %v, want %v",
			err,
			ErrNoUsableAirports,
		)
	}
	if catalog != nil ||
		report.InputCount != 1 ||
		report.AcceptedCount != 0 ||
		report.ExcludedCount != 1 {
		t.Fatalf(
			"unexpected result: catalog=%#v report=%#v",
			catalog,
			report,
		)
	}
}

func airportFixture(
	icaoCode string,
	iataCode string,
	name string,
	latitude float64,
	longitude float64,
) airport.Airport {
	return airport.Airport{
		ICAOCode:   icaoCode,
		IATACode:   iataCode,
		Name:       name,
		City:       "City",
		Country:    "Country",
		Latitude:   latitude,
		Longitude:  longitude,
		ElevationM: 10,
		Timezone:   "UTC",
	}
}
