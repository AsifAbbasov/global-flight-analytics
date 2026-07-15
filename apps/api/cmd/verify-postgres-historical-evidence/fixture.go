package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5"
)

const (
	fixtureOriginICAO      = "UBBB"
	fixtureDestinationICAO = "UGTB"
	fixtureVersion         = "historical-evidence-verification-v1"
)

type fixtureFlight struct {
	FlightID      string
	TrajectoryID  string
	RouteRecordID string

	ICAO24   string
	Callsign string

	StartTime time.Time
	EndTime   time.Time

	RouteJSON        []byte
	RouteFingerprint string
}

type fixtureObservation struct {
	ID       string
	FlightID string
	ICAO24   string
	Callsign string

	ObservedAt time.Time
	Latitude   float64
	Longitude  float64
}

type evidenceFixture struct {
	Marker     string
	SourceName string

	Flights      []fixtureFlight
	Observations []fixtureObservation

	FlightIDs      []string
	TrajectoryIDs  []string
	ObservationIDs []string
	RouteRecordIDs []string
}

func buildEvidenceFixture(
	schedule evidenceSchedule,
) (evidenceFixture, error) {
	seed := fmt.Sprintf(
		"%s|%d",
		schedule.AsOfTime.Format(
			time.RFC3339Nano,
		),
		os.Getpid(),
	)
	markerDigest := sha256.Sum256(
		[]byte(seed),
	)
	marker := hex.EncodeToString(
		markerDigest[:8],
	)

	fixture := evidenceFixture{
		Marker: "historical-evidence-" +
			marker,
		SourceName: "historical_evidence_verification_" +
			marker,
		Flights:      make([]fixtureFlight, 0, 7),
		Observations: make([]fixtureObservation, 0, 15),
	}

	offsets := []time.Duration{
		-3*time.Hour - 30*time.Minute,
		-2*time.Hour - 30*time.Minute,
		-1*time.Hour - 45*time.Minute,
		-1*time.Hour - 15*time.Minute,
		-45 * time.Minute,
		-30 * time.Minute,
		-15 * time.Minute,
	}

	for index, offset := range offsets {
		startTime := schedule.ClosedBoundary.Add(
			offset,
		)
		endTime := startTime.Add(
			10 * time.Minute,
		)
		flightID := deterministicUUID(
			fixture.Marker +
				fmt.Sprintf(
					"|flight|%d",
					index,
				),
		)
		trajectoryID := deterministicUUID(
			fixture.Marker +
				fmt.Sprintf(
					"|trajectory|%d",
					index,
				),
		)
		icao24 := fmt.Sprintf(
			"e%05x",
			index+1,
		)
		callsign := fmt.Sprintf(
			"HEV%03d",
			index+1,
		)

		routeFingerprint := fingerprint(
			fixture.Marker +
				fmt.Sprintf(
					"|route|%d",
					index,
				),
		)
		routeRecordID := "route-record-" +
			hashHex(
				fixture.Marker+
					fmt.Sprintf(
						"|route-record|%d",
						index,
					),
			)

		routeResult := routecontract.Result{
			SchemaVersion: routecontract.SchemaVersionV1,
			Status:        routecontract.RouteStatusComplete,

			TrajectoryID: trajectoryID,
			FlightID:     flightID,
			ICAO24:       icao24,
			Callsign:     callsign,

			Window: routecontract.RouteWindow{
				StartTime: startTime,
				EndTime:   endTime,
				AsOfTime:  endTime,
			},
			Origin: &routecontract.EndpointInference{
				Role: routecontract.
					EndpointRoleOrigin,
				Airport: routecontract.AirportReference{
					ICAOCode:  fixtureOriginICAO,
					IATACode:  "GYD",
					Name:      "Verification Origin",
					City:      "Baku",
					Country:   "Azerbaijan",
					Latitude:  40.4675,
					Longitude: 50.0467,
				},
				Confidence: routecontract.Confidence{
					Score: 0.95,
					Level: routecontract.
						ConfidenceLevelHigh,
					EvidenceCount: 1,
				},
			},
			Destination: &routecontract.EndpointInference{
				Role: routecontract.
					EndpointRoleDestination,
				Airport: routecontract.AirportReference{
					ICAOCode: fixtureDestinationICAO,
					IATACode: "TBS",
					Name: "Verification " +
						"Destination",
					City:      "Tbilisi",
					Country:   "Georgia",
					Latitude:  41.6692,
					Longitude: 44.9547,
				},
				Confidence: routecontract.Confidence{
					Score: 0.94,
					Level: routecontract.
						ConfidenceLevelHigh,
					EvidenceCount: 1,
				},
			},
			Summary: routecontract.RouteSummary{
				GreatCircleDistanceKM: 448.5,
				SameAirport:           false,
			},
			Confidence: routecontract.Confidence{
				Score: 0.90,
				Level: routecontract.
					ConfidenceLevelHigh,
				EvidenceCount: 2,
			},
			Provenance: routecontract.Provenance{
				ResolverVersion:     fixtureVersion,
				InputFingerprint:    routeFingerprint,
				TrajectoryUpdatedAt: endTime,
				SourceNames: []string{
					fixture.SourceName,
				},
			},
			GeneratedAt: endTime,
		}

		routeJSON, err := json.Marshal(
			routeResult,
		)
		if err != nil {
			return evidenceFixture{},
				fmt.Errorf(
					"marshal route fixture %d: %w",
					index,
					err,
				)
		}

		fixture.Flights = append(
			fixture.Flights,
			fixtureFlight{
				FlightID:         flightID,
				TrajectoryID:     trajectoryID,
				RouteRecordID:    routeRecordID,
				ICAO24:           icao24,
				Callsign:         callsign,
				StartTime:        startTime,
				EndTime:          endTime,
				RouteJSON:        routeJSON,
				RouteFingerprint: routeFingerprint,
			},
		)
		fixture.FlightIDs = append(
			fixture.FlightIDs,
			flightID,
		)
		fixture.TrajectoryIDs = append(
			fixture.TrajectoryIDs,
			trajectoryID,
		)
		fixture.RouteRecordIDs = append(
			fixture.RouteRecordIDs,
			routeRecordID,
		)
	}

	bucketStarts := []time.Time{
		schedule.ClosedBoundary.Add(
			-4 * time.Hour,
		),
		schedule.ClosedBoundary.Add(
			-3 * time.Hour,
		),
		schedule.ClosedBoundary.Add(
			-2 * time.Hour,
		),
		schedule.ClosedBoundary.Add(
			-time.Hour,
		),
	}
	observationCounts := []int{
		2,
		3,
		4,
		6,
	}
	flightIndexes := [][]int{
		{0},
		{1},
		{2, 3},
		{4, 5, 6},
	}

	observationSequence := 0
	for bucketIndex, bucketStart := range bucketStarts {
		count := observationCounts[bucketIndex]
		indexes := flightIndexes[bucketIndex]
		for index := 0; index < count; index++ {
			flight := fixture.Flights[indexes[index%len(indexes)]]
			observedAt := bucketStart.Add(
				time.Duration(index+1) *
					time.Hour /
					time.Duration(count+1),
			)
			observationID := deterministicUUID(
				fixture.Marker +
					fmt.Sprintf(
						"|observation|%d",
						observationSequence,
					),
			)

			fixture.Observations = append(
				fixture.Observations,
				fixtureObservation{
					ID:         observationID,
					FlightID:   flight.FlightID,
					ICAO24:     flight.ICAO24,
					Callsign:   flight.Callsign,
					ObservedAt: observedAt,
					Latitude: 40.0 +
						float64(observationSequence)/
							100,
					Longitude: 49.0 +
						float64(observationSequence)/
							100,
				},
			)
			fixture.ObservationIDs = append(
				fixture.ObservationIDs,
				observationID,
			)
			observationSequence++
		}
	}

	return fixture, nil
}

func insertEvidenceFixture(
	ctx context.Context,
	tx pgx.Tx,
	fixture evidenceFixture,
) error {
	for _, flight := range fixture.Flights {
		if _, err := tx.Exec(
			ctx,
			`
				INSERT INTO flights (
					id,
					callsign,
					first_seen_at,
					last_seen_at,
					status,
					created_at,
					updated_at
				)
				VALUES (
					$1::uuid,
					$2,
					$3,
					$4,
					'completed',
					$3,
					$4
				);
			`,
			flight.FlightID,
			flight.Callsign,
			flight.StartTime,
			flight.EndTime,
		); err != nil {
			return fmt.Errorf(
				"insert flight %s: %w",
				flight.FlightID,
				err,
			)
		}

		if _, err := tx.Exec(
			ctx,
			`
				INSERT INTO flight_trajectories (
					id,
					flight_id,
					icao24,
					callsign,
					start_time,
					end_time,
					duration_seconds,
					segment_count,
					point_count,
					coverage_gap_count,
					quality_score,
					source_name,
					metadata_json,
					created_at,
					updated_at
				)
				VALUES (
					$1::uuid,
					$2::uuid,
					$3,
					$4,
					$5,
					$6,
					$7,
					1,
					2,
					0,
					0.95,
					$8,
					$9::jsonb,
					$5,
					$6
				);
			`,
			flight.TrajectoryID,
			flight.FlightID,
			flight.ICAO24,
			flight.Callsign,
			flight.StartTime,
			flight.EndTime,
			int(
				flight.EndTime.Sub(
					flight.StartTime,
				).Seconds(),
			),
			fixture.SourceName,
			`{"verification":true}`,
		); err != nil {
			return fmt.Errorf(
				"insert trajectory %s: %w",
				flight.TrajectoryID,
				err,
			)
		}

		if _, err := tx.Exec(
			ctx,
			`
				INSERT INTO flight_route_results (
					id,
					trajectory_id,
					schema_version,
					as_of_time,
					as_of_time_unix_nano,
					input_fingerprint,
					route_status,
					confidence_level,
					validation_warning_count,
					route_json,
					stored_at,
					stored_at_unix_nano
				)
				VALUES (
					$1,
					$2::uuid,
					'route-intelligence-v1',
					$3,
					$4,
					$5,
					'complete',
					'high',
					0,
					$6::jsonb,
					$3,
					$4
				);
			`,
			flight.RouteRecordID,
			flight.TrajectoryID,
			flight.EndTime,
			flight.EndTime.UnixNano(),
			flight.RouteFingerprint,
			flight.RouteJSON,
		); err != nil {
			return fmt.Errorf(
				"insert route result %s: %w",
				flight.RouteRecordID,
				err,
			)
		}
	}

	for _, observation := range fixture.Observations {
		if _, err := tx.Exec(
			ctx,
			`
				INSERT INTO flight_states (
					id,
					flight_id,
					icao24,
					callsign,
					latitude,
					longitude,
					barometric_altitude_status,
					geometric_altitude_status,
					on_ground,
					observed_at,
					source_name,
					created_at
				)
				VALUES (
					$1::uuid,
					$2::uuid,
					$3,
					$4,
					$5,
					$6,
					'unavailable',
					'unavailable',
					false,
					$7,
					$8,
					$7
				);
			`,
			observation.ID,
			observation.FlightID,
			observation.ICAO24,
			observation.Callsign,
			observation.Latitude,
			observation.Longitude,
			observation.ObservedAt,
			fixture.SourceName,
		); err != nil {
			return fmt.Errorf(
				"insert observation %s: %w",
				observation.ID,
				err,
			)
		}
	}

	return nil
}

func deterministicUUID(
	seed string,
) string {
	digest := sha256.Sum256(
		[]byte(seed),
	)
	raw := append(
		[]byte(nil),
		digest[:16]...,
	)
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(raw)
	return encoded[0:8] + "-" +
		encoded[8:12] + "-" +
		encoded[12:16] + "-" +
		encoded[16:20] + "-" +
		encoded[20:32]
}

func fingerprint(
	seed string,
) string {
	return "sha256:" + hashHex(seed)
}

func hashHex(
	seed string,
) string {
	digest := sha256.Sum256(
		[]byte(seed),
	)
	return hex.EncodeToString(
		digest[:],
	)
}
