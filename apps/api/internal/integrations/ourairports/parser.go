package ourairports

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

const SourceName = "ourairports"

var requiredHeaders = []string{
	"ident",
	"name",
	"latitude_deg",
	"longitude_deg",
	"elevation_ft",
	"iso_country",
	"municipality",
	"icao_code",
	"iata_code",
}

func ParseAirportsCSV(
	reader io.Reader,
	syncedAt time.Time,
) ([]airport.ImportRecord, error) {
	return parseAirportsCSV(
		reader,
		syncedAt,
		nil,
	)
}

func ParseAirportsCSVForCountryCodes(
	reader io.Reader,
	syncedAt time.Time,
	countryCodes []string,
) ([]airport.ImportRecord, error) {
	allowedCountryCodes, err := buildAllowedCountryCodeSet(
		countryCodes,
	)
	if err != nil {
		return nil, err
	}

	return parseAirportsCSV(
		reader,
		syncedAt,
		allowedCountryCodes,
	)
}

func parseAirportsCSV(
	reader io.Reader,
	syncedAt time.Time,
	allowedCountryCodes map[string]struct{},
) ([]airport.ImportRecord, error) {
	if reader == nil {
		return nil, errors.New(
			"OurAirports CSV reader is required",
		)
	}

	csvReader := csv.NewReader(reader)

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf(
			"read OurAirports CSV header: %w",
			err,
		)
	}

	indexes, err := buildHeaderIndexes(header)
	if err != nil {
		return nil, err
	}

	csvReader.FieldsPerRecord = len(header)

	items := make(
		[]airport.ImportRecord,
		0,
	)

	rowNumber := 1

	for {
		rowNumber++

		row, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf(
				"read OurAirports CSV row %d: %w",
				rowNumber,
				err,
			)
		}

		if !isCountryAllowed(
			row,
			indexes,
			allowedCountryCodes,
		) {
			continue
		}

		item, err := parseAirportRow(
			row,
			indexes,
			syncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse OurAirports CSV row %d: %w",
				rowNumber,
				err,
			)
		}

		items = append(
			items,
			item,
		)
	}

	return items, nil
}

func buildAllowedCountryCodeSet(
	countryCodes []string,
) (map[string]struct{}, error) {
	allowedCountryCodes := make(
		map[string]struct{},
		len(countryCodes),
	)

	for _, countryCode := range countryCodes {
		normalizedCountryCode := strings.ToUpper(
			strings.TrimSpace(countryCode),
		)

		if normalizedCountryCode == "" {
			continue
		}

		allowedCountryCodes[normalizedCountryCode] = struct{}{}
	}

	if len(allowedCountryCodes) == 0 {
		return nil, ErrCountryCodesRequired
	}

	return allowedCountryCodes, nil
}

func isCountryAllowed(
	row []string,
	indexes map[string]int,
	allowedCountryCodes map[string]struct{},
) bool {
	if allowedCountryCodes == nil {
		return true
	}

	countryCode := strings.ToUpper(
		csvValue(
			row,
			indexes,
			"iso_country",
		),
	)

	_, exists := allowedCountryCodes[countryCode]

	return exists
}

func buildHeaderIndexes(
	header []string,
) (map[string]int, error) {
	indexes := make(
		map[string]int,
		len(header),
	)

	for index, name := range header {
		indexes[strings.TrimSpace(name)] = index
	}

	for _, requiredHeader := range requiredHeaders {
		if _, exists := indexes[requiredHeader]; !exists {
			return nil, fmt.Errorf(
				"required CSV header %q is missing",
				requiredHeader,
			)
		}
	}

	return indexes, nil
}

func parseAirportRow(
	row []string,
	indexes map[string]int,
	syncedAt time.Time,
) (airport.ImportRecord, error) {
	sourceIdent := csvValue(
		row,
		indexes,
		"ident",
	)
	if sourceIdent == "" {
		return airport.ImportRecord{},
			errors.New("airport ident is required")
	}

	name := csvValue(
		row,
		indexes,
		"name",
	)
	if name == "" {
		return airport.ImportRecord{},
			errors.New("airport name is required")
	}

	latitude, err := parseRequiredFloat(
		csvValue(
			row,
			indexes,
			"latitude_deg",
		),
		"latitude_deg",
	)
	if err != nil {
		return airport.ImportRecord{}, err
	}

	longitude, err := parseRequiredFloat(
		csvValue(
			row,
			indexes,
			"longitude_deg",
		),
		"longitude_deg",
	)
	if err != nil {
		return airport.ImportRecord{}, err
	}

	elevationFT, err := parseOptionalInt(
		csvValue(
			row,
			indexes,
			"elevation_ft",
		),
		"elevation_ft",
	)
	if err != nil {
		return airport.ImportRecord{}, err
	}

	return airport.ImportRecord{
		SourceIdent: sourceIdent,
		ICAOCode: strings.ToUpper(
			csvValue(
				row,
				indexes,
				"icao_code",
			),
		),
		IATACode: strings.ToUpper(
			csvValue(
				row,
				indexes,
				"iata_code",
			),
		),
		Name: name,
		City: csvValue(
			row,
			indexes,
			"municipality",
		),
		SourceCountryCode: strings.ToUpper(
			csvValue(
				row,
				indexes,
				"iso_country",
			),
		),
		Latitude:     latitude,
		Longitude:    longitude,
		ElevationFT:  elevationFT,
		SourceName:   SourceName,
		LastSyncedAt: syncedAt.UTC(),
	}, nil
}

func csvValue(
	row []string,
	indexes map[string]int,
	name string,
) string {
	return strings.TrimSpace(
		row[indexes[name]],
	)
}

func parseRequiredFloat(
	value string,
	fieldName string,
) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf(
			"%s is required",
			fieldName,
		)
	}

	parsed, err := strconv.ParseFloat(
		value,
		64,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s: %w",
			fieldName,
			err,
		)
	}

	return parsed, nil
}

func parseOptionalInt(
	value string,
	fieldName string,
) (*int, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf(
			"parse %s: %w",
			fieldName,
			err,
		)
	}

	return &parsed, nil
}
