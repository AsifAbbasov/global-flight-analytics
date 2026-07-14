package routeresolver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type Resolver struct {
	now                         func() time.Time
	partialConfidenceFactor     float64
	sameAirportConfidenceFactor float64
}

func New(config Config) (*Resolver, error) {
	partialConfidenceFactor :=
		config.PartialConfidenceFactor
	if partialConfidenceFactor == 0 {
		partialConfidenceFactor =
			DefaultPartialConfidenceFactor
	}
	if !finiteRatio(partialConfidenceFactor) {
		return nil,
			ErrInvalidPartialConfidenceFactor
	}

	sameAirportConfidenceFactor :=
		config.SameAirportConfidenceFactor
	if sameAirportConfidenceFactor == 0 {
		sameAirportConfidenceFactor =
			DefaultSameAirportConfidenceFactor
	}
	if !finiteRatio(sameAirportConfidenceFactor) {
		return nil,
			ErrInvalidSameAirportConfidenceFactor
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Resolver{
		now:                         now,
		partialConfidenceFactor:     partialConfidenceFactor,
		sameAirportConfidenceFactor: sameAirportConfidenceFactor,
	}, nil
}

func (resolver *Resolver) Resolve(
	ctx context.Context,
	input Input,
) (Resolution, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Resolution{}, err
	}
	if err := validateEndpointEvidence(
		input.Origin,
		routecontract.EndpointRoleOrigin,
	); err != nil {
		return Resolution{}, err
	}
	if err := validateEndpointEvidence(
		input.Destination,
		routecontract.EndpointRoleDestination,
	); err != nil {
		return Resolution{}, err
	}

	sourceNames, err := normalizeSourceNames(
		input.SourceNames,
		input.Origin,
		input.Destination,
	)
	if err != nil {
		return Resolution{}, err
	}

	normalizedInput := input
	normalizedInput.TrajectoryID =
		strings.TrimSpace(input.TrajectoryID)
	normalizedInput.IdentityKey =
		strings.TrimSpace(input.IdentityKey)
	normalizedInput.FlightID =
		strings.TrimSpace(input.FlightID)
	normalizedInput.AircraftID =
		strings.TrimSpace(input.AircraftID)
	normalizedInput.ICAO24 = strings.ToUpper(
		strings.TrimSpace(input.ICAO24),
	)
	normalizedInput.Callsign =
		strings.TrimSpace(input.Callsign)
	normalizedInput.Window = routecontract.RouteWindow{
		StartTime: normalizeUTC(input.Window.StartTime),
		EndTime:   normalizeUTC(input.Window.EndTime),
		AsOfTime:  normalizeUTC(input.Window.AsOfTime),
	}
	normalizedInput.TrajectoryUpdatedAt =
		normalizeUTC(input.TrajectoryUpdatedAt)
	normalizedInput.SourceNames = sourceNames

	generatedAt := resolver.now().UTC()
	if !normalizedInput.Window.AsOfTime.IsZero() &&
		generatedAt.Before(
			normalizedInput.Window.AsOfTime,
		) {
		return Resolution{},
			ErrGeneratedBeforeAsOfTime
	}

	origin := selectedEndpoint(
		normalizedInput.Origin,
	)
	destination := selectedEndpoint(
		normalizedInput.Destination,
	)

	status := routeStatus(origin, destination)
	summary := routeSummary(origin, destination)
	confidence := resolver.routeConfidence(
		origin,
		destination,
		summary.SameAirport,
	)
	limitations := routeLimitations(
		normalizedInput.Origin,
		normalizedInput.Destination,
		summary.SameAirport,
	)

	inputFingerprint, err := resolverInputFingerprint(
		normalizedInput,
		resolver.partialConfidenceFactor,
		resolver.sameAirportConfidenceFactor,
	)
	if err != nil {
		return Resolution{}, err
	}

	result := routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        status,
		TrajectoryID:  normalizedInput.TrajectoryID,
		IdentityKey:   normalizedInput.IdentityKey,
		FlightID:      normalizedInput.FlightID,
		AircraftID:    normalizedInput.AircraftID,
		ICAO24:        normalizedInput.ICAO24,
		Callsign:      normalizedInput.Callsign,
		Window:        normalizedInput.Window,
		Origin:        origin,
		Destination:   destination,
		Summary:       summary,
		Confidence:    confidence,
		Limitations:   limitations,
		Provenance: routecontract.Provenance{
			ResolverVersion:  Version,
			InputFingerprint: inputFingerprint,
			TrajectoryUpdatedAt: normalizedInput.
				TrajectoryUpdatedAt,
			SourceNames: append(
				[]string(nil),
				sourceNames...,
			),
		},
		GeneratedAt: generatedAt,
	}

	report := routecontract.Validate(result)
	if report.Status !=
		routecontract.ValidationStatusValid {
		return Resolution{},
			&ContractValidationError{
				Report: report.Clone(),
			}
	}

	return Resolution{
		Version:    Version,
		Result:     result.Clone(),
		Validation: report.Clone(),
	}.Clone(), nil
}

func selectedEndpoint(
	result endpointevidence.Result,
) *routecontract.EndpointInference {
	if result.Status !=
		endpointevidence.SelectionStatusSelected ||
		result.Endpoint == nil {
		return nil
	}

	return result.Clone().Endpoint
}

func routeStatus(
	origin *routecontract.EndpointInference,
	destination *routecontract.EndpointInference,
) routecontract.RouteStatus {
	switch {
	case origin != nil && destination != nil:
		return routecontract.RouteStatusComplete
	case origin != nil || destination != nil:
		return routecontract.RouteStatusPartial
	default:
		return routecontract.RouteStatusUnavailable
	}
}

func routeSummary(
	origin *routecontract.EndpointInference,
	destination *routecontract.EndpointInference,
) routecontract.RouteSummary {
	if origin == nil || destination == nil {
		return routecontract.RouteSummary{}
	}

	sameAirport := strings.EqualFold(
		origin.Airport.ICAOCode,
		destination.Airport.ICAOCode,
	)
	if sameAirport {
		return routecontract.RouteSummary{
			SameAirport: true,
		}
	}

	return routecontract.RouteSummary{
		GreatCircleDistanceKM: greatCircleDistanceKM(
			origin.Airport.Latitude,
			origin.Airport.Longitude,
			destination.Airport.Latitude,
			destination.Airport.Longitude,
		),
		SameAirport: false,
	}
}

func (resolver *Resolver) routeConfidence(
	origin *routecontract.EndpointInference,
	destination *routecontract.EndpointInference,
	sameAirport bool,
) routecontract.Confidence {
	var score float64
	var evidenceCount int
	reasons := make(
		[]routecontract.ConfidenceReason,
		0,
		2,
	)

	switch {
	case origin != nil && destination != nil:
		score = math.Min(
			origin.Confidence.Score,
			destination.Confidence.Score,
		)
		evidenceCount =
			len(origin.Evidence) +
				len(destination.Evidence)
		reasons = append(
			reasons,
			routecontract.ConfidenceReason{
				Code:         "both_route_endpoints_available",
				Message:      "Both origin and destination are supported by selected endpoint evidence.",
				Contribution: score,
			},
		)
		if sameAirport {
			penalty := score *
				(1 -
					resolver.
						sameAirportConfidenceFactor)
			score = clampUnit(
				score *
					resolver.
						sameAirportConfidenceFactor,
			)
			reasons = append(
				reasons,
				routecontract.ConfidenceReason{
					Code:         "same_airport_candidate",
					Message:      "Origin and destination resolve to the same airport, so complete-route confidence is reduced.",
					Contribution: -penalty,
				},
			)
		}

	case origin != nil || destination != nil:
		selected := origin
		if selected == nil {
			selected = destination
		}
		evidenceCount = len(selected.Evidence)
		score = clampUnit(
			selected.Confidence.Score *
				resolver.
					partialConfidenceFactor,
		)
		reasons = append(
			reasons,
			routecontract.ConfidenceReason{
				Code:         "single_route_endpoint_available",
				Message:      "Only one route endpoint is selected, so overall route confidence is reduced.",
				Contribution: score,
			},
		)

	default:
		reasons = append(
			reasons,
			routecontract.ConfidenceReason{
				Code:         "no_route_endpoints_available",
				Message:      "No route endpoint passed selection requirements.",
				Contribution: 0,
			},
		)
	}

	sort.SliceStable(
		reasons,
		func(left int, right int) bool {
			return reasons[left].Code <
				reasons[right].Code
		},
	)

	return routecontract.Confidence{
		Score: score,
		Level: routecontract.
			ConfidenceLevelForScore(score),
		EvidenceCount: evidenceCount,
		Reasons:       reasons,
	}
}

func routeLimitations(
	origin endpointevidence.Result,
	destination endpointevidence.Result,
	sameAirport bool,
) []routecontract.Limitation {
	limitations := []routecontract.Limitation{
		{
			Code:    "probable_route_only",
			Message: "Route endpoints are inferred and are not filed or operational flight-plan data.",
			Scope:   "route",
		},
	}

	if origin.Status !=
		endpointevidence.SelectionStatusSelected {
		limitations = append(
			limitations,
			endpointStatusLimitation(
				routecontract.EndpointRoleOrigin,
				origin.Status,
			),
		)
	}
	if destination.Status !=
		endpointevidence.SelectionStatusSelected {
		limitations = append(
			limitations,
			endpointStatusLimitation(
				routecontract.
					EndpointRoleDestination,
				destination.Status,
			),
		)
	}
	if (origin.Status ==
		endpointevidence.SelectionStatusSelected) !=
		(destination.Status ==
			endpointevidence.SelectionStatusSelected) {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code:    "route_incomplete",
				Message: "Only one route endpoint is selected, so the route is incomplete.",
				Scope:   "route",
			},
		)
	}
	if destination.Status ==
		endpointevidence.SelectionStatusSelected {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code:    "destination_not_planned_destination",
				Message: "The selected destination reflects persisted trajectory evidence and may not be the planned destination.",
				Scope:   "route",
			},
		)
	}
	if sameAirport {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code:    "same_airport_route_candidate",
				Message: "Origin and destination resolve to the same airport, which may represent a local flight, incomplete coverage, or uncertain inference.",
				Scope:   "route",
			},
		)
	}

	sort.SliceStable(
		limitations,
		func(left int, right int) bool {
			return limitations[left].Code <
				limitations[right].Code
		},
	)

	return limitations
}

func endpointStatusLimitation(
	role routecontract.EndpointRole,
	status endpointevidence.SelectionStatus,
) routecontract.Limitation {
	return routecontract.Limitation{
		Code: fmt.Sprintf(
			"%s_endpoint_%s",
			role,
			status,
		),
		Message: fmt.Sprintf(
			"The %s endpoint selection status is %s.",
			role,
			status,
		),
		Scope: "route",
	}
}

func resolverInputFingerprint(
	input Input,
	partialConfidenceFactor float64,
	sameAirportConfidenceFactor float64,
) (string, error) {
	payload := struct {
		TrajectoryID string `json:"trajectory_id"`
		IdentityKey  string `json:"identity_key"`
		FlightID     string `json:"flight_id"`
		AircraftID   string `json:"aircraft_id"`
		ICAO24       string `json:"icao24"`
		Callsign     string `json:"callsign"`

		StartTime              string   `json:"start_time"`
		EndTime                string   `json:"end_time"`
		AsOfTime               string   `json:"as_of_time"`
		TrajectoryUpdatedAt    string   `json:"trajectory_updated_at"`
		OriginFingerprint      string   `json:"origin_fingerprint"`
		DestinationFingerprint string   `json:"destination_fingerprint"`
		SourceNames            []string `json:"source_names"`

		PartialConfidenceFactor     float64 `json:"partial_confidence_factor"`
		SameAirportConfidenceFactor float64 `json:"same_airport_confidence_factor"`
	}{
		TrajectoryID: input.TrajectoryID,
		IdentityKey:  input.IdentityKey,
		FlightID:     input.FlightID,
		AircraftID:   input.AircraftID,
		ICAO24:       input.ICAO24,
		Callsign:     input.Callsign,
		StartTime: input.Window.StartTime.Format(
			time.RFC3339Nano,
		),
		EndTime: input.Window.EndTime.Format(
			time.RFC3339Nano,
		),
		AsOfTime: input.Window.AsOfTime.Format(
			time.RFC3339Nano,
		),
		TrajectoryUpdatedAt: input.TrajectoryUpdatedAt.Format(
			time.RFC3339Nano,
		),
		OriginFingerprint:      input.Origin.InputFingerprint,
		DestinationFingerprint: input.Destination.InputFingerprint,
		SourceNames: append(
			[]string(nil),
			input.SourceNames...,
		),
		PartialConfidenceFactor:     partialConfidenceFactor,
		SameAirportConfidenceFactor: sameAirportConfidenceFactor,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
