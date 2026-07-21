package postgres

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5/pgtype"
)

const internationalFootInMeters = 0.3048

type airportRowScanner interface {
	Scan(destinations ...any) error
}

type airportRecord struct {
	ID   string
	Item airport.Airport
}

func scanAirportRecord(scanner airportRowScanner) (airportRecord, error) {
	var record airportRecord
	var elevationFeet pgtype.Int4

	if err := scanner.Scan(
		&record.ID,
		&record.Item.ICAOCode,
		&record.Item.IATACode,
		&record.Item.Name,
		&record.Item.City,
		&record.Item.Country,
		&record.Item.Latitude,
		&record.Item.Longitude,
		&elevationFeet,
		&record.Item.Timezone,
		&record.Item.Description,
	); err != nil {
		return airportRecord{}, err
	}

	applyAirportElevationDatabaseValue(&record.Item, elevationFeet)
	return record, nil
}

func applyAirportElevationDatabaseValue(
	item *airport.Airport,
	elevationFeet pgtype.Int4,
) {
	item.ElevationM = 0
	item.ElevationAvailable = false
	if !elevationFeet.Valid {
		return
	}

	item.ElevationM = feetToMeters(float64(elevationFeet.Int32))
	item.ElevationAvailable = true
}

func feetToMeters(feet float64) float64 {
	return feet * internationalFootInMeters
}
