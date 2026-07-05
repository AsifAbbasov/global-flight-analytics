package ourairports

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseAirportsCSVForCountryCodesFiltersBeforeParsing(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
US-INVALID,Invalid Airport,not-a-number,also-invalid,,US,Example City,,
`

	syncedAt := time.Date(
		2026,
		time.July,
		4,
		14,
		30,
		0,
		0,
		time.UTC,
	)

	items, err := ParseAirportsCSVForCountryCodes(
		strings.NewReader(csvData),
		syncedAt,
		[]string{"AZ"},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(items) != 1 {
		t.Fatalf(
			"expected 1 airport, got %d",
			len(items),
		)
	}

	if items[0].SourceIdent != "UBBB" {
		t.Fatalf(
			"expected source ident UBBB, got %s",
			items[0].SourceIdent,
		)
	}

	if items[0].SourceCountryCode != "AZ" {
		t.Fatalf(
			"expected country code AZ, got %s",
			items[0].SourceCountryCode,
		)
	}
}

func TestParseAirportsCSVForCountryCodesRejectsMissingCountryCodes(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSVForCountryCodes(
		strings.NewReader(csvData),
		time.Time{},
		nil,
	)

	if !errors.Is(
		err,
		ErrCountryCodesRequired,
	) {
		t.Fatalf(
			"expected country codes required error, got %v",
			err,
		)
	}
}
