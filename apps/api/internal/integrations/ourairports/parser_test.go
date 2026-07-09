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

func TestParseAirportsCSVAcceptsCoordinateBoundaryValues(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
NORTH,North Boundary Airport,90,180,0,AZ,North,TEST,NTH
SOUTH,South Boundary Airport,-90,-180,0,AZ,South,TES2,STH
`

	items, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err != nil {
		t.Fatalf(
			"parse coordinate boundary values: %v",
			err,
		)
	}

	if len(items) != 2 {
		t.Fatalf(
			"expected 2 airports, got %d",
			len(items),
		)
	}

	if items[0].Latitude != 90 {
		t.Fatalf(
			"expected north latitude 90, got %v",
			items[0].Latitude,
		)
	}

	if items[0].Longitude != 180 {
		t.Fatalf(
			"expected east longitude 180, got %v",
			items[0].Longitude,
		)
	}

	if items[1].Latitude != -90 {
		t.Fatalf(
			"expected south latitude -90, got %v",
			items[1].Latitude,
		)
	}

	if items[1].Longitude != -180 {
		t.Fatalf(
			"expected west longitude -180, got %v",
			items[1].Longitude,
		)
	}
}

func TestParseAirportsCSVRejectsLatitudeAboveMaximum(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,90.0001,50.0467,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected latitude above maximum error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"latitude_deg",
	) {
		t.Fatalf(
			"expected latitude_deg error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsLatitudeBelowMinimum(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,-90.0001,50.0467,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected latitude below minimum error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"latitude_deg",
	) {
		t.Fatalf(
			"expected latitude_deg error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsLongitudeAboveMaximum(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,180.0001,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected longitude above maximum error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"longitude_deg",
	) {
		t.Fatalf(
			"expected longitude_deg error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsLongitudeBelowMinimum(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,-180.0001,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected longitude below minimum error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"longitude_deg",
	) {
		t.Fatalf(
			"expected longitude_deg error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsNonFiniteLatitude(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,NaN,50.0467,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected non-finite latitude error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"latitude_deg",
	) {
		t.Fatalf(
			"expected latitude_deg error, got %v",
			err,
		)
	}
}

func TestParseAirportsCSVRejectsNonFiniteLongitude(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,+Inf,10,AZ,Baku,UBBB,GYD
`

	_, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Time{},
	)
	if err == nil {
		t.Fatal(
			"expected non-finite longitude error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"longitude_deg",
	) {
		t.Fatalf(
			"expected longitude_deg error, got %v",
			err,
		)
	}
}
