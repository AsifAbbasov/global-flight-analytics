package projectionread

import (
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func projectionReadTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func projectionReadTrajectory(
	id string,
	endTime time.Time,
) trajectory.FlightTrajectory {
	endTime = endTime.UTC()
	startTime := endTime.Add(
		-10 * time.Minute,
	)
	points := []trajectory.TrackPoint4D{
		projectionReadPoint(
			"point-a",
			startTime,
			40.40,
			49.80,
		),
		projectionReadPoint(
			"point-b",
			startTime.Add(5*time.Minute),
			40.45,
			49.90,
		),
		projectionReadPoint(
			"point-c",
			endTime,
			40.50,
			50.00,
		),
	}

	return trajectory.FlightTrajectory{
		ID: id,
		IdentityKey: "flight-identity-" +
			strings.Repeat("a", 64),
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		FlightID:   "6b57d421-9f75-4f1b-931d-d4e658515d92",
		AircraftID: "a20eef16-c12c-41fd-870e-cd5a814ef3ad",
		ICAO24:     "4A1234",
		Callsign:   "AHY123",
		StartTime:  startTime,
		EndTime:    endTime,
		DurationSeconds: int64(
			endTime.Sub(startTime) /
				time.Second,
		),
		SegmentCount:     1,
		PointCount:       len(points),
		CoverageGapCount: 0,
		QualityScore:     0.95,
		SourceName:       "projection-read-test",
		Points:           points,
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "segment-a",
				TrajectoryID:   id,
				FlightID:       "6b57d421-9f75-4f1b-931d-d4e658515d92",
				AircraftID:     "a20eef16-c12c-41fd-870e-cd5a814ef3ad",
				ICAO24:         "4A1234",
				Callsign:       "AHY123",
				SequenceNumber: 0,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore: 0.95,
				StartTime:    startTime,
				EndTime:      endTime,
				DurationSeconds: int64(
					endTime.Sub(startTime) /
						time.Second,
				),
				StartLatitude:  40.40,
				StartLongitude: 49.80,
				EndLatitude:    40.50,
				EndLongitude:   50.00,
				PointCount:     len(points),
				SourceName:     "projection-read-test",
				CreatedAt:      endTime,
			},
		},
		CoverageGaps: []trajectory.CoverageGap{},
		CreatedAt:    endTime,
		UpdatedAt:    endTime,
	}
}

func projectionReadPoint(
	id string,
	observedAt time.Time,
	latitude float64,
	longitude float64,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		ID:                       id,
		FlightStateID:            id,
		FlightID:                 "6b57d421-9f75-4f1b-931d-d4e658515d92",
		AircraftID:               "a20eef16-c12c-41fd-870e-cd5a814ef3ad",
		ICAO24:                   "4A1234",
		Callsign:                 "AHY123",
		Latitude:                 latitude,
		Longitude:                longitude,
		BarometricAltitudeM:      10000,
		BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:       10100,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              220,
		HeadingDegrees:           270,
		VerticalRateMPS:          0,
		OnGround:                 false,
		OriginCountry:            "Azerbaijan",
		ObservedAt:               observedAt.UTC(),
		SourceName:               "projection-read-test",
	}
}

func projectionReadCompleteRoute(
	current trajectory.FlightTrajectory,
	asOfTime time.Time,
) routecontract.Result {
	asOfTime = asOfTime.UTC()
	originEvidence :=
		projectionReadRouteEvidence(
			"origin",
			asOfTime,
		)
	destinationEvidence :=
		projectionReadRouteEvidence(
			"destination",
			asOfTime,
		)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  current.ID,
		IdentityKey:   current.IdentityKey,
		FlightID:      current.FlightID,
		AircraftID:    current.AircraftID,
		ICAO24:        current.ICAO24,
		Callsign:      current.Callsign,
		Window: routecontract.RouteWindow{
			StartTime: current.StartTime.UTC(),
			EndTime:   current.EndTime.UTC(),
			AsOfTime:  asOfTime,
		},
		Origin: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleOrigin,
			Airport: projectionReadAirport(
				"UBBB",
				"GYD",
				"Heydar Aliyev International Airport",
				40.4675,
				50.0467,
			),
			DistanceKM: 5,
			Confidence: projectionReadRouteConfidence(
				0.90,
				1,
				"origin_confidence",
			),
			Evidence: []routecontract.Evidence{
				originEvidence,
			},
			Limitations: []routecontract.Limitation{},
		},
		Destination: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleDestination,
			Airport: projectionReadAirport(
				"LTBA",
				"ISL",
				"Istanbul Ataturk Airport",
				40.9769,
				28.8146,
			),
			DistanceKM: 6,
			Confidence: projectionReadRouteConfidence(
				0.90,
				1,
				"destination_confidence",
			),
			Evidence: []routecontract.Evidence{
				destinationEvidence,
			},
			Limitations: []routecontract.Limitation{},
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 1760,
			SameAirport:           false,
		},
		Confidence: projectionReadRouteConfidence(
			0.90,
			2,
			"route_confidence",
		),
		Limitations: []routecontract.Limitation{},
		Provenance: routecontract.Provenance{
			ResolverVersion: "projection-read-test-route-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
			TrajectoryUpdatedAt: current.UpdatedAt.UTC(),
			SourceNames: []string{
				"projection-read-test",
			},
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func projectionReadAirport(
	icao string,
	iata string,
	name string,
	latitude float64,
	longitude float64,
) routecontract.AirportReference {
	return routecontract.AirportReference{
		ICAOCode:   icao,
		IATACode:   iata,
		Name:       name,
		City:       "",
		Country:    "",
		Latitude:   latitude,
		Longitude:  longitude,
		ElevationM: 0,
		Timezone:   "UTC",
	}
}

func projectionReadRouteEvidence(
	summary string,
	asOfTime time.Time,
) routecontract.Evidence {
	return routecontract.Evidence{
		Type: routecontract.
			EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    "projection-read-test",
		SourceVersion: "projection-read-test-v1",
		Score:         0.90,
		Weight:        1,
		ObservedAt:    asOfTime.UTC(),
		Summary:       summary,
		Attributes:    []routecontract.EvidenceAttribute{},
	}
}

func projectionReadRouteConfidence(
	score float64,
	evidenceCount int,
	code string,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score: score,
		Level: routecontract.
			ConfidenceLevelForScore(
				score,
			),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Projection read test confidence.",
				Contribution: score,
			},
		},
	}
}

func projectionReadHistory(
	asOfTime time.Time,
) projectionroutefrequency.HistorySummary {
	asOfTime = asOfTime.UTC()
	summary :=
		projectionroutefrequency.HistorySummary{
			RouteKey: "UBBB>LTBA",
			WindowStart: asOfTime.Add(
				-180 * 24 * time.Hour,
			),
			WindowEnd:              asOfTime,
			AsOfTime:               asOfTime,
			ObservationCount:       10,
			DistinctFlightCount:    8,
			DistinctDayCount:       7,
			RecentObservationCount: 4,
			LastObservedAt: asOfTime.Add(
				-24 * time.Hour,
			),
			SourceNames: []string{
				DefaultSourceName,
			},
		}
	summary.InputFingerprint =
		routeHistoryFingerprint(
			summary,
		)

	return summary
}
