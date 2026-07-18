package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchbenchmark"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchdataset"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
	fmt.Println(
		"Open aviation research evidence verification: PASS",
	)
}

func run() error {
	if err := verifyTransponderEvidence(); err != nil {
		return err
	}
	if err := verifyDatasetBoundaries(); err != nil {
		return err
	}
	for _, plan := range researchbenchmark.DefaultPlans() {
		if err := researchbenchmark.Validate(plan); err != nil {
			return fmt.Errorf(
				"validate benchmark plan %s: %w",
				plan.ID,
				err,
			)
		}
	}
	return nil
}

func verifyTransponderEvidence() error {
	start := time.Date(
		2026,
		time.July,
		18,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	evidence, err := transponderalert.Build(
		[]flightstate.FlightState{
			{
				ICAO24:     "4k001",
				SquawkCode: "7700",
				ObservedAt: start,
				SourceName: "opensky",
			},
			{
				ICAO24:     "4k001",
				SquawkCode: "7700",
				ObservedAt: start.Add(20 * time.Second),
				SourceName: "opensky",
			},
		},
		start.Add(time.Minute),
	)
	if err != nil {
		return fmt.Errorf(
			"build transponder evidence: %w",
			err,
		)
	}
	if len(evidence) != 1 ||
		evidence[0].MaximumClaimStrength !=
			"observed_transponder_code_only" {
		return errors.New(
			"transponder evidence claim boundary failed",
		)
	}
	return nil
}

func verifyDatasetBoundaries() error {
	decision, err := researchdataset.EvaluateSourceBoundary(
		researchdataset.IDADSC,
	)
	if err != nil {
		return err
	}
	if decision.Usable() {
		return errors.New(
			"ADS-C satellite dataset was not blocked",
		)
	}

	manifest := researchdataset.Manifest{
		DatasetID: researchdataset.IDWeeklyStateVectors,
		Version:   "verification-v1",
		Files: []researchdataset.File{
			{
				Name:      "bounded-sample.avro",
				Format:    "avro",
				SizeBytes: 1024,
				SHA256:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
		TotalBytes:           1024,
		MaximumRecords:       100,
		RegionFilter:         "AZ,GE,AM,TR",
		OfflineOnly:          true,
		ProductionDependency: false,
		LicenseReviewed:      true,
		AttributionProvided:  true,
		PreparedAt:           time.Now().UTC(),
	}
	allowed, err := researchdataset.ValidateManifest(
		manifest,
	)
	if err != nil {
		return fmt.Errorf(
			"validate bounded manifest: %w",
			err,
		)
	}
	if !allowed.Allowed {
		return errors.New(
			"bounded research manifest was not allowed",
		)
	}

	manifest.DatasetID =
		researchdataset.IDTrinoSnapshot2026
	manifest.SelectedTables = []string{
		"readsb_adsc_sv",
	}
	_, err = researchdataset.ValidateManifest(
		manifest,
	)
	if !errors.Is(
		err,
		researchdataset.ErrBlockedTable,
	) {
		return errors.New(
			"blocked ADS-C table was not rejected",
		)
	}
	return nil
}
