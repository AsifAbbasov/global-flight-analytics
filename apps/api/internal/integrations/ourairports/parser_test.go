package ourairports

import (
	"strings"
	"testing"
	"time"
)

func TestParseAirportsCSV(t *testing.T) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
AZ-0001,Example Local Airport,40.1000,49.9000,,AZ,Example City,,
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

	items, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		syncedAt,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(items) != 2 {
		t.Fatalf(
			"expected 2 airports, got %d",
			len(items),
		)
	}

	first := items[0]

	if first.SourceIdent != "UBBB" {
		t.Fatalf(
			"expected source ident UBBB, got %s",
			first.SourceIdent,
		)
	}

	if first.ICAOCode != "UBBB" {
		t.Fatalf(
			"expected ICAO code UBBB, got %s",
			first.ICAOCode,
		)
	}

	if first.IATACode != "GYD" {
		t.Fatalf(
			"expected IATA code GYD, got %s",
			first.IATACode,
		)
	}

	if first.SourceCountryCode != "AZ" {
		t.Fatalf(
			"expected country code AZ, got %s",
			first.SourceCountryCode,
		)
	}

	if first.ElevationFT == nil {
		t.Fatal(
			"expected elevation in feet",
		)
	}

	if *first.ElevationFT != 10 {
		t.Fatalf(
			"expected elevation 10 feet, got %d",
			*first.ElevationFT,
		)
	}

	if first.SourceName != SourceName {
		t.Fatalf(
			"expected source name %s, got %s",
			SourceName,
			first.SourceName,
		)
	}

	if !first.LastSyncedAt.Equal(syncedAt) {
		t.Fatalf(
			"expected synced time %s, got %s",
			syncedAt,
			first.LastSyncedAt,
		)
	}

	second := items[1]

	if second.SourceIdent != "AZ-0001" {
		t.Fatalf(
			"expected source ident AZ-0001, got %s",
			second.SourceIdent,
		)
	}

	if second.ICAOCode != "" {
		t.Fatalf(
			"expected empty ICAO code, got %s",
			second.ICAOCode,
		)
	}

	if second.IATACode != "" {
		t.Fatalf(
			"expected empty IATA code, got %s",
			second.IATACode,
		)
	}

	if second.ElevationFT != nil {
		t.Fatalf(
			"expected missing elevation, got %d",
			*second.ElevationFT,
		)
	}
}

func TestParseAirportsCSVRejectsInvalidLatitude(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,invalid,50.0467,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected invalid latitude error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"latitude_deg",
	) {
		t.Fatalf(
			"expected latitude error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsMissingRequiredHeader(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected missing header error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"iata_code",
	) {
		t.Fatalf(
			"expected missing iata_code error, got %v",
			err,
		)
	}
}
