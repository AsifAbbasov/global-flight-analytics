package ourairports

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseAirportsCSVRejectsPublicationAtomicallyAfterValidRow(t *testing.T) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
BROKEN,Broken Airport,invalid,50.0000,0,AZ,Baku,UBBX,BRK
`

	items, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Now().UTC(),
	)
	if len(items) != 0 {
		t.Fatalf("partial publication escaped: %d rows", len(items))
	}
	if !errors.Is(err, ErrAtomicPublicationRejected) {
		t.Fatalf("error = %v, want %v", err, ErrAtomicPublicationRejected)
	}

	var publicationError *AtomicPublicationError
	if !errors.As(err, &publicationError) {
		t.Fatalf("error = %T, want *AtomicPublicationError", err)
	}
	if publicationError.RowNumber != 3 {
		t.Fatalf("row number = %d, want 3", publicationError.RowNumber)
	}
}

func TestParseAirportsCSVRejectsStructurallyMalformedRowAtomically(t *testing.T) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
BROKEN,Too,Few,Columns
`

	items, err := ParseAirportsCSV(
		strings.NewReader(csvData),
		time.Now().UTC(),
	)
	if len(items) != 0 {
		t.Fatalf("partial publication escaped: %d rows", len(items))
	}
	if !errors.Is(err, ErrAtomicPublicationRejected) {
		t.Fatalf("error = %v, want %v", err, ErrAtomicPublicationRejected)
	}
}
