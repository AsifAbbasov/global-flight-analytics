package routeresolver

import (
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func validateEndpointEvidence(
	result endpointevidence.Result,
	expectedRole routecontract.EndpointRole,
) error {
	if result.Version != endpointevidence.Version ||
		result.Role != expectedRole ||
		!fingerprintPattern.MatchString(
			result.InputFingerprint,
		) ||
		result.CandidateCount < 0 ||
		result.SelectedCandidateRank < 0 ||
		!finiteRatio(result.SelectedCandidateScore) ||
		!finiteRatio(result.RunnerUpCandidateScore) ||
		!finiteRatio(result.CandidateScoreGap) {
		return ErrInvalidEndpointEvidence
	}

	switch result.Status {
	case endpointevidence.SelectionStatusUnavailable:
		if result.Endpoint != nil ||
			result.CandidateCount != 0 ||
			result.SelectedCandidateRank != 0 {
			return ErrInvalidEndpointEvidence
		}
	case endpointevidence.SelectionStatusInsufficient:
		if result.Endpoint != nil ||
			result.CandidateCount < 1 ||
			result.SelectedCandidateRank < 1 {
			return ErrInvalidEndpointEvidence
		}
	case endpointevidence.SelectionStatusAmbiguous:
		if result.Endpoint != nil ||
			result.CandidateCount < 2 ||
			result.SelectedCandidateRank < 1 {
			return ErrInvalidEndpointEvidence
		}
	case endpointevidence.SelectionStatusSelected:
		if result.Endpoint == nil ||
			result.CandidateCount < 1 ||
			result.SelectedCandidateRank < 1 ||
			result.Endpoint.Role != expectedRole {
			return ErrInvalidEndpointEvidence
		}
	default:
		return ErrInvalidEndpointEvidence
	}

	return nil
}

func normalizeSourceNames(
	items []string,
	origin endpointevidence.Result,
	destination endpointevidence.Result,
) ([]string, error) {
	unique := make(map[string]struct{})

	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		unique[normalized] = struct{}{}
	}

	for _, result := range []endpointevidence.Result{
		origin,
		destination,
	} {
		if result.Endpoint == nil {
			continue
		}
		for _, evidence := range result.Endpoint.Evidence {
			normalized := strings.TrimSpace(
				evidence.SourceName,
			)
			if normalized == "" {
				continue
			}
			unique[normalized] = struct{}{}
		}
	}

	if len(unique) == 0 {
		return nil, ErrSourceNamesRequired
	}

	result := make([]string, 0, len(unique))
	for item := range unique {
		result = append(result, item)
	}
	sortStrings(result)

	return result, nil
}

func normalizeUTC(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}

	return value.UTC()
}

func sortStrings(items []string) {
	for index := 1; index < len(items); index++ {
		value := items[index]
		position := index - 1
		for position >= 0 && items[position] > value {
			items[position+1] = items[position]
			position--
		}
		items[position+1] = value
	}
}
