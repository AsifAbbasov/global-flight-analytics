package projectionneighbors

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const RouteScopeVersion = "projection-neighbor-route-scope-v1"

type RouteScopeMode string

const (
	RouteScopeUniform  RouteScopeMode = "uniform_route_scope"
	RouteScopeExplicit RouteScopeMode = "explicit_candidate_routes"
)

func (mode RouteScopeMode) IsKnown() bool {
	switch mode {
	case RouteScopeUniform, RouteScopeExplicit:
		return true
	default:
		return false
	}
}

type RouteKey struct {
	OriginICAO      string
	DestinationICAO string
}

func (key RouteKey) Equal(other RouteKey) bool {
	return normalizeRouteKey(key) == normalizeRouteKey(other)
}

type CandidateRouteEvidence struct {
	TrajectoryID     string
	Route            RouteKey
	SourceName       string
	InputFingerprint string
}

type RouteScope struct {
	Version string
	Mode    RouteScopeMode
	Route   RouteKey

	SourceName string
	Candidates []CandidateRouteEvidence

	InputFingerprint string
}

func UniformRouteScope(
	originICAO string,
	destinationICAO string,
	sourceName string,
) RouteScope {
	scope := RouteScope{
		Version: RouteScopeVersion,
		Mode:    RouteScopeUniform,
		Route: normalizeRouteKey(RouteKey{
			OriginICAO:      originICAO,
			DestinationICAO: destinationICAO,
		}),
		SourceName: strings.TrimSpace(sourceName),
	}
	scope.InputFingerprint = routeScopeFingerprint(scope)
	return scope
}

func CandidateRoute(
	trajectoryID string,
	originICAO string,
	destinationICAO string,
	sourceName string,
) CandidateRouteEvidence {
	evidence := CandidateRouteEvidence{
		TrajectoryID: strings.TrimSpace(trajectoryID),
		Route: normalizeRouteKey(RouteKey{
			OriginICAO:      originICAO,
			DestinationICAO: destinationICAO,
		}),
		SourceName: strings.TrimSpace(sourceName),
	}
	evidence.InputFingerprint = candidateRouteFingerprint(evidence)
	return evidence
}

func ExplicitRouteScope(
	originICAO string,
	destinationICAO string,
	sourceName string,
	candidates []CandidateRouteEvidence,
) RouteScope {
	scope := RouteScope{
		Version: RouteScopeVersion,
		Mode:    RouteScopeExplicit,
		Route: normalizeRouteKey(RouteKey{
			OriginICAO:      originICAO,
			DestinationICAO: destinationICAO,
		}),
		SourceName: strings.TrimSpace(sourceName),
		Candidates: make(
			[]CandidateRouteEvidence,
			0,
			len(candidates),
		),
	}
	for _, candidate := range candidates {
		normalized := normalizeCandidateRouteEvidence(candidate)
		normalized.InputFingerprint = candidateRouteFingerprint(normalized)
		scope.Candidates = append(scope.Candidates, normalized)
	}
	scope.InputFingerprint = routeScopeFingerprint(scope)
	return scope
}

func (scope RouteScope) Clone() RouteScope {
	cloned := scope
	cloned.Candidates = append(
		[]CandidateRouteEvidence(nil),
		scope.Candidates...,
	)
	return cloned
}

func (scope RouteScope) ValidateForCandidates(
	candidates []trajectory.FlightTrajectory,
) error {
	_, err := prepareRouteScope(scope, candidates)
	return err
}

type routeScopeIndex struct {
	scope               RouteScope
	evidenceByCandidate map[string]CandidateRouteEvidence
}

func prepareRouteScope(
	scope RouteScope,
	candidates []trajectory.FlightTrajectory,
) (routeScopeIndex, error) {
	normalized := normalizeRouteScope(scope)
	if err := validateRouteScopeIdentity(scope, normalized); err != nil {
		return routeScopeIndex{}, err
	}

	candidateIDs := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidateID := strings.TrimSpace(candidate.ID)
		if candidateID != "" {
			candidateIDs[candidateID] = struct{}{}
		}
	}

	index := routeScopeIndex{
		scope: normalized,
		evidenceByCandidate: make(
			map[string]CandidateRouteEvidence,
			len(candidateIDs),
		),
	}

	switch normalized.Mode {
	case RouteScopeUniform:
		if len(normalized.Candidates) != 0 {
			return routeScopeIndex{}, fmt.Errorf(
				"uniform route scope must not contain explicit candidate routes",
			)
		}
		for candidateID := range candidateIDs {
			index.evidenceByCandidate[candidateID] = CandidateRoute(
				candidateID,
				normalized.Route.OriginICAO,
				normalized.Route.DestinationICAO,
				normalized.SourceName,
			)
		}
	case RouteScopeExplicit:
		for _, candidate := range normalized.Candidates {
			if _, exists := index.evidenceByCandidate[candidate.TrajectoryID]; exists {
				return routeScopeIndex{}, fmt.Errorf(
					"duplicate route evidence for candidate %q",
					candidate.TrajectoryID,
				)
			}
			if _, exists := candidateIDs[candidate.TrajectoryID]; !exists {
				return routeScopeIndex{}, fmt.Errorf(
					"route evidence references unknown candidate %q",
					candidate.TrajectoryID,
				)
			}
			if err := validateCandidateRouteEvidence(candidate); err != nil {
				return routeScopeIndex{}, fmt.Errorf(
					"validate route evidence for candidate %q: %w",
					candidate.TrajectoryID,
					err,
				)
			}
			index.evidenceByCandidate[candidate.TrajectoryID] = candidate
		}
		for candidateID := range candidateIDs {
			if _, exists := index.evidenceByCandidate[candidateID]; !exists {
				return routeScopeIndex{}, fmt.Errorf(
					"route evidence is missing for candidate %q",
					candidateID,
				)
			}
		}
	default:
		return routeScopeIndex{}, fmt.Errorf(
			"route scope mode is invalid: %q",
			normalized.Mode,
		)
	}

	return index, nil
}

func validateRouteScopeIdentity(
	original RouteScope,
	normalized RouteScope,
) error {
	if normalized.Version != RouteScopeVersion {
		return fmt.Errorf(
			"route scope version is invalid: %q",
			normalized.Version,
		)
	}
	if !normalized.Mode.IsKnown() {
		return fmt.Errorf(
			"route scope mode is invalid: %q",
			normalized.Mode,
		)
	}
	if err := validateRouteKey(normalized.Route); err != nil {
		return err
	}
	if normalized.SourceName == "" {
		return fmt.Errorf("route scope source name is required")
	}
	if !fingerprintPattern.MatchString(original.InputFingerprint) ||
		original.InputFingerprint != routeScopeFingerprint(normalized) {
		return fmt.Errorf("route scope input fingerprint is invalid")
	}
	return nil
}

func validateCandidateRouteEvidence(
	evidence CandidateRouteEvidence,
) error {
	if strings.TrimSpace(evidence.TrajectoryID) == "" {
		return fmt.Errorf("candidate trajectory identifier is required")
	}
	if err := validateRouteKey(evidence.Route); err != nil {
		return err
	}
	if strings.TrimSpace(evidence.SourceName) == "" {
		return fmt.Errorf("candidate route evidence source name is required")
	}
	if !fingerprintPattern.MatchString(evidence.InputFingerprint) ||
		evidence.InputFingerprint != candidateRouteFingerprint(evidence) {
		return fmt.Errorf("candidate route evidence fingerprint is invalid")
	}
	return nil
}

var icaoCodePattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)

func validateRouteKey(key RouteKey) error {
	if !icaoCodePattern.MatchString(key.OriginICAO) ||
		!icaoCodePattern.MatchString(key.DestinationICAO) {
		return fmt.Errorf(
			"route key requires four-character origin and destination ICAO codes",
		)
	}
	return nil
}

func normalizeRouteScope(scope RouteScope) RouteScope {
	normalized := scope.Clone()
	normalized.Version = strings.TrimSpace(normalized.Version)
	normalized.Route = normalizeRouteKey(normalized.Route)
	normalized.SourceName = strings.TrimSpace(normalized.SourceName)
	for index := range normalized.Candidates {
		normalized.Candidates[index] = normalizeCandidateRouteEvidence(
			normalized.Candidates[index],
		)
	}
	return normalized
}

func normalizeRouteKey(key RouteKey) RouteKey {
	return RouteKey{
		OriginICAO: strings.ToUpper(strings.TrimSpace(key.OriginICAO)),
		DestinationICAO: strings.ToUpper(
			strings.TrimSpace(key.DestinationICAO),
		),
	}
}

func normalizeCandidateRouteEvidence(
	evidence CandidateRouteEvidence,
) CandidateRouteEvidence {
	evidence.TrajectoryID = strings.TrimSpace(evidence.TrajectoryID)
	evidence.Route = normalizeRouteKey(evidence.Route)
	evidence.SourceName = strings.TrimSpace(evidence.SourceName)
	return evidence
}

func candidateRouteFingerprint(
	evidence CandidateRouteEvidence,
) string {
	normalized := normalizeCandidateRouteEvidence(evidence)
	digest := sha256.New()
	writeFingerprintString(digest, RouteScopeVersion+":candidate")
	writeFingerprintString(digest, normalized.TrajectoryID)
	writeFingerprintString(digest, normalized.Route.OriginICAO)
	writeFingerprintString(digest, normalized.Route.DestinationICAO)
	writeFingerprintString(digest, normalized.SourceName)
	return fingerprintPrefix + hex.EncodeToString(digest.Sum(nil))
}

func routeScopeFingerprint(scope RouteScope) string {
	normalized := normalizeRouteScope(scope)
	digest := sha256.New()
	writeFingerprintString(digest, RouteScopeVersion+":scope")
	writeFingerprintString(digest, string(normalized.Mode))
	writeFingerprintString(digest, normalized.Route.OriginICAO)
	writeFingerprintString(digest, normalized.Route.DestinationICAO)
	writeFingerprintString(digest, normalized.SourceName)

	candidates := append(
		[]CandidateRouteEvidence(nil),
		normalized.Candidates...,
	)
	sort.SliceStable(candidates, func(left int, right int) bool {
		if candidates[left].TrajectoryID != candidates[right].TrajectoryID {
			return candidates[left].TrajectoryID < candidates[right].TrajectoryID
		}
		return candidates[left].InputFingerprint <
			candidates[right].InputFingerprint
	})
	writeFingerprintInt(digest, len(candidates))
	for _, candidate := range candidates {
		writeFingerprintString(digest, candidate.InputFingerprint)
	}
	return fingerprintPrefix + hex.EncodeToString(digest.Sum(nil))
}
