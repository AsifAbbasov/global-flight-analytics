package projectionread

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const unavailableRouteResolverVersion = "projection-read-route-unavailable-v1"

func unavailableRoute(
	current trajectory.FlightTrajectory,
	asOfTime time.Time,
	generatedAt time.Time,
) routecontract.Result {
	asOfTime = asOfTime.UTC()
	generatedAt = generatedAt.UTC()

	startTime := current.StartTime.UTC()
	if startTime.IsZero() ||
		startTime.After(asOfTime) {
		startTime = asOfTime
	}

	endTime := current.EndTime.UTC()
	if len(current.Points) > 0 {
		endTime = current.Points[len(current.Points)-1].
			ObservedAt.UTC()
	}
	if endTime.IsZero() ||
		endTime.After(asOfTime) {
		endTime = asOfTime
	}
	if endTime.Before(startTime) {
		startTime = endTime
	}

	trajectoryUpdatedAt :=
		current.UpdatedAt.UTC()
	if trajectoryUpdatedAt.IsZero() ||
		trajectoryUpdatedAt.After(asOfTime) {
		trajectoryUpdatedAt = endTime
	}

	fingerprint := unavailableRouteFingerprint(
		current.ID,
		asOfTime,
		endTime,
	)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusUnavailable,

		TrajectoryID: strings.TrimSpace(
			current.ID,
		),
		IdentityKey: strings.TrimSpace(
			current.IdentityKey,
		),
		FlightID: strings.TrimSpace(
			current.FlightID,
		),
		AircraftID: strings.TrimSpace(
			current.AircraftID,
		),
		ICAO24: strings.ToUpper(
			strings.TrimSpace(
				current.ICAO24,
			),
		),
		Callsign: strings.TrimSpace(
			current.Callsign,
		),

		Window: routecontract.RouteWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Summary: routecontract.RouteSummary{},
		Confidence: routecontract.Confidence{
			Score:         0,
			Level:         routecontract.ConfidenceLevelNone,
			EvidenceCount: 0,
			Reasons:       []routecontract.ConfidenceReason{},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "route_intelligence_not_materialized",
				Message: "No Route Intelligence result was materialized at or before the requested analytical time.",
				Scope:   "route",
			},
		},
		Provenance: routecontract.Provenance{
			ResolverVersion:     unavailableRouteResolverVersion,
			InputFingerprint:    fingerprint,
			TrajectoryUpdatedAt: trajectoryUpdatedAt,
			SourceNames: []string{
				DefaultSourceName,
			},
		},
		GeneratedAt: generatedAt,
	}
}

func unavailableRouteFingerprint(
	trajectoryID string,
	asOfTime time.Time,
	endTime time.Time,
) string {
	digest := sha256.Sum256(
		[]byte(
			fmt.Sprintf(
				"%s|%s|%s|%s",
				unavailableRouteResolverVersion,
				strings.TrimSpace(
					trajectoryID,
				),
				asOfTime.UTC().Format(
					time.RFC3339Nano,
				),
				endTime.UTC().Format(
					time.RFC3339Nano,
				),
			),
		),
	)

	return "sha256:" +
		hex.EncodeToString(
			digest[:],
		)
}
