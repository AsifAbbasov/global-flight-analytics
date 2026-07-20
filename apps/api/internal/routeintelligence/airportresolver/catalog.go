package airportresolver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	icaoCodePattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
	iataCodePattern = regexp.MustCompile(
		`^[A-Z0-9]{3}$`,
	)
)

type Catalog struct {
	airports    []routecontract.AirportReference
	fingerprint string
}

func NewCatalog(
	items []airport.Airport,
) (*Catalog, CatalogBuildReport, error) {
	report := CatalogBuildReport{
		Version:    CatalogVersion,
		InputCount: len(items),
	}
	exclusionCounts := make(
		map[ExclusionReason]int,
	)

	normalized := make(
		[]routecontract.AirportReference,
		0,
		len(items),
	)
	for _, item := range items {
		reference, reason, valid := normalizeAirport(
			item,
		)
		if !valid {
			exclusionCounts[reason]++
			continue
		}

		normalized = append(normalized, reference)
	}

	sort.SliceStable(
		normalized,
		func(left int, right int) bool {
			return canonicalAirportKey(normalized[left]) <
				canonicalAirportKey(normalized[right])
		},
	)

	unique := make(
		[]routecontract.AirportReference,
		0,
		len(normalized),
	)
	for index := 0; index < len(normalized); {
		next := index + 1
		for next < len(normalized) &&
			normalized[next].ICAOCode ==
				normalized[index].ICAOCode {
			next++
		}

		unique = append(unique, normalized[index])
		duplicateCount := next - index - 1
		if duplicateCount > 0 {
			exclusionCounts[ExclusionReasonDuplicateICAOCode] +=
				duplicateCount
		}

		index = next
	}

	if len(unique) == 0 {
		report.ExcludedCount = report.InputCount
		report.Exclusions = exclusionSummaries(
			exclusionCounts,
		)

		return nil, report.Clone(), ErrNoUsableAirports
	}

	fingerprint, err := fingerprintAirports(unique)
	if err != nil {
		return nil, report.Clone(), err
	}

	report.AcceptedCount = len(unique)
	report.ExcludedCount =
		report.InputCount - report.AcceptedCount
	report.Exclusions = exclusionSummaries(
		exclusionCounts,
	)
	report.Fingerprint = fingerprint

	catalog := &Catalog{
		airports: append(
			[]routecontract.AirportReference(nil),
			unique...,
		),
		fingerprint: fingerprint,
	}

	return catalog, report.Clone(), nil
}

func (catalog *Catalog) Version() string {
	return CatalogVersion
}

func (catalog *Catalog) Fingerprint() string {
	if catalog == nil {
		return ""
	}

	return catalog.fingerprint
}

func (catalog *Catalog) Size() int {
	if catalog == nil {
		return 0
	}

	return len(catalog.airports)
}

func (catalog *Catalog) Airports() []routecontract.AirportReference {
	if catalog == nil {
		return nil
	}

	return append(
		[]routecontract.AirportReference(nil),
		catalog.airports...,
	)
}

func normalizeAirport(
	item airport.Airport,
) (
	routecontract.AirportReference,
	ExclusionReason,
	bool,
) {
	icaoCode := strings.ToUpper(
		strings.TrimSpace(item.ICAOCode),
	)
	if !icaoCodePattern.MatchString(icaoCode) {
		return routecontract.AirportReference{},
			ExclusionReasonInvalidICAOCode,
			false
	}

	iataCode := strings.ToUpper(
		strings.TrimSpace(item.IATACode),
	)
	if iataCode != "" &&
		!iataCodePattern.MatchString(iataCode) {
		return routecontract.AirportReference{},
			ExclusionReasonInvalidIATACode,
			false
	}

	name := strings.TrimSpace(item.Name)
	if name == "" {
		return routecontract.AirportReference{},
			ExclusionReasonMissingName,
			false
	}

	if !validLatitude(item.Latitude) ||
		!validLongitude(item.Longitude) {
		return routecontract.AirportReference{},
			ExclusionReasonInvalidCoordinates,
			false
	}

	elevationM, elevationStatus, elevationAvailable :=
		airport.ResolveElevation(
			item.ElevationM,
			item.ElevationAvailable,
		)
	if elevationStatus == airport.ElevationStatusInvalid {
		return routecontract.AirportReference{},
			ExclusionReasonInvalidElevation,
			false
	}

	return routecontract.AirportReference{
		ICAOCode:           icaoCode,
		IATACode:           iataCode,
		Name:               name,
		City:               strings.TrimSpace(item.City),
		Country:            strings.TrimSpace(item.Country),
		Latitude:           normalizeSignedZero(item.Latitude),
		Longitude:          normalizeSignedZero(item.Longitude),
		ElevationM:         normalizeSignedZero(elevationM),
		ElevationAvailable: elevationAvailable,
		Timezone:           strings.TrimSpace(item.Timezone),
	}, "", true
}

func exclusionSummaries(
	counts map[ExclusionReason]int,
) []ExclusionSummary {
	reasons := make(
		[]ExclusionReason,
		0,
		len(counts),
	)
	for reason, count := range counts {
		if count <= 0 {
			continue
		}

		reasons = append(reasons, reason)
	}

	sort.Slice(
		reasons,
		func(left int, right int) bool {
			return reasons[left] < reasons[right]
		},
	)

	result := make(
		[]ExclusionSummary,
		0,
		len(reasons),
	)
	for _, reason := range reasons {
		result = append(
			result,
			ExclusionSummary{
				Reason: reason,
				Count:  counts[reason],
			},
		)
	}

	return result
}

func canonicalAirportKey(
	item routecontract.AirportReference,
) string {
	return strings.Join(
		[]string{
			item.ICAOCode,
			item.IATACode,
			item.Name,
			item.City,
			item.Country,
			strconv.FormatFloat(
				item.Latitude,
				'f',
				-1,
				64,
			),
			strconv.FormatFloat(
				item.Longitude,
				'f',
				-1,
				64,
			),
			strconv.FormatFloat(
				item.ElevationM,
				'f',
				-1,
				64,
			),
			strconv.FormatBool(item.ElevationAvailable),
			item.Timezone,
		},
		"\x00",
	)
}

func fingerprintAirports(
	items []routecontract.AirportReference,
) (string, error) {
	encoded, err := json.Marshal(items)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func normalizeSignedZero(
	value float64,
) float64 {
	if value == 0 {
		return 0
	}

	return value
}
