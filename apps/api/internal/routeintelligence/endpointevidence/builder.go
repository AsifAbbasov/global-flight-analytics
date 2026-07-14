package endpointevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type Builder struct {
	minimumSelectionScore    float64
	minimumCandidateScoreGap float64
}

func New(config Config) (*Builder, error) {
	minimumSelectionScore := config.MinimumSelectionScore
	if minimumSelectionScore == 0 {
		minimumSelectionScore =
			DefaultMinimumSelectionScore
	}
	if !finiteRatio(minimumSelectionScore) {
		return nil, ErrInvalidMinimumSelectionScore
	}

	minimumCandidateScoreGap :=
		config.MinimumCandidateScoreGap
	if minimumCandidateScoreGap == 0 {
		minimumCandidateScoreGap =
			DefaultMinimumCandidateScoreGap
	}
	if !finiteRatio(minimumCandidateScoreGap) {
		return nil, ErrInvalidMinimumCandidateScoreGap
	}

	return &Builder{
		minimumSelectionScore:    minimumSelectionScore,
		minimumCandidateScoreGap: minimumCandidateScoreGap,
	}, nil
}

func (builder *Builder) Build(
	ctx context.Context,
	input Input,
) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := validateInput(input); err != nil {
		return Result{}, err
	}

	fingerprint, err := inputFingerprint(
		input,
		builder.minimumSelectionScore,
		builder.minimumCandidateScoreGap,
	)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Version:          Version,
		Role:             input.Candidates.Role,
		CandidateCount:   len(input.Candidates.Candidates),
		InputFingerprint: fingerprint,
	}

	if len(input.Candidates.Candidates) == 0 {
		result.Status = SelectionStatusUnavailable
		result.Limitations = []routecontract.Limitation{
			{
				Code:    "airport_candidate_unavailable",
				Message: "No airport candidate is available within the configured geographic radius.",
				Scope:   string(input.Candidates.Role),
			},
		}

		return result.Clone(), nil
	}

	topCandidate := input.Candidates.Candidates[0]
	topBreakdown := scoreCandidate(
		topCandidate,
		input.TrajectoryQuality,
		input.SegmentStatus,
		input.SegmentPointCount,
	)
	result.SelectedCandidateRank = topCandidate.Rank
	result.SelectedCandidateScore = topBreakdown.Total
	result.CandidateScoreGap = topBreakdown.Total

	var runnerUpBreakdown scoreBreakdown
	if len(input.Candidates.Candidates) > 1 {
		runnerUpBreakdown = scoreCandidate(
			input.Candidates.Candidates[1],
			input.TrajectoryQuality,
			input.SegmentStatus,
			input.SegmentPointCount,
		)
		result.RunnerUpCandidateScore =
			runnerUpBreakdown.Total
		result.CandidateScoreGap = clamp01(
			topBreakdown.Total -
				runnerUpBreakdown.Total,
		)
	}

	switch {
	case topBreakdown.Total <
		builder.minimumSelectionScore:
		result.Status = SelectionStatusInsufficient
		result.Limitations = []routecontract.Limitation{
			{
				Code: "endpoint_confidence_insufficient",
				Message: fmt.Sprintf(
					"The leading airport candidate score %.3f is below the required threshold %.3f.",
					topBreakdown.Total,
					builder.minimumSelectionScore,
				),
				Scope: string(input.Candidates.Role),
			},
		}

		return result.Clone(), nil

	case len(input.Candidates.Candidates) > 1 &&
		result.CandidateScoreGap <
			builder.minimumCandidateScoreGap:
		result.Status = SelectionStatusAmbiguous
		result.Limitations = []routecontract.Limitation{
			{
				Code: "airport_candidate_ambiguous",
				Message: fmt.Sprintf(
					"The leading candidate score gap %.3f is below the required separation %.3f.",
					result.CandidateScoreGap,
					builder.minimumCandidateScoreGap,
				),
				Scope: string(input.Candidates.Role),
			},
		}

		return result.Clone(), nil
	}

	result.Status = SelectionStatusSelected
	result.Endpoint = buildEndpoint(
		input,
		topCandidate,
		topBreakdown,
		result.CandidateScoreGap,
	)
	result.Limitations = append(
		[]routecontract.Limitation(nil),
		result.Endpoint.Limitations...,
	)

	return result.Clone(), nil
}

func buildEndpoint(
	input Input,
	candidate airportresolver.Candidate,
	breakdown scoreBreakdown,
	scoreGap float64,
) *routecontract.EndpointInference {
	role := input.Candidates.Role
	limitations := endpointLimitations(input)

	reasons := []routecontract.ConfidenceReason{
		{
			Code: fmt.Sprintf(
				"%s_airport_proximity",
				role,
			),
			Message: fmt.Sprintf(
				"The selected airport is %.3f kilometres from the persisted trajectory endpoint.",
				candidate.DistanceKM,
			),
			Contribution: breakdown.ProximityContribution,
		},
		{
			Code: "endpoint_point_evidence",
			Message: fmt.Sprintf(
				"The endpoint segment contains %d persisted point(s).",
				input.SegmentPointCount,
			),
			Contribution: breakdown.PointEvidenceContribution,
		},
		{
			Code: "endpoint_segment_status",
			Message: fmt.Sprintf(
				"The endpoint segment status is %s.",
				input.SegmentStatus,
			),
			Contribution: breakdown.SegmentStatusContribution,
		},
		{
			Code: "trajectory_quality_evidence",
			Message: fmt.Sprintf(
				"The trajectory quality score is %.3f.",
				input.TrajectoryQuality,
			),
			Contribution: breakdown.TrajectoryQualityContribution,
		},
	}
	sort.SliceStable(
		reasons,
		func(left int, right int) bool {
			return reasons[left].Code <
				reasons[right].Code
		},
	)

	evidence := routecontract.Evidence{
		Type: routecontract.
			EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    "trajectory_endpoint",
		SourceVersion: input.Candidates.Version,
		Score:         breakdown.Total,
		Weight:        1,
		ObservedAt:    input.ObservedAt,
		Summary: fmt.Sprintf(
			"The %s trajectory endpoint supports airport %s as the leading geographic candidate.",
			role,
			candidate.Airport.ICAOCode,
		),
		Attributes: evidenceAttributes(
			input,
			candidate,
			scoreGap,
		),
	}

	return &routecontract.EndpointInference{
		Role:       role,
		Airport:    candidate.Airport,
		DistanceKM: candidate.DistanceKM,
		Confidence: routecontract.Confidence{
			Score: breakdown.Total,
			Level: routecontract.
				ConfidenceLevelForScore(
					breakdown.Total,
				),
			EvidenceCount: 1,
			Reasons:       reasons,
		},
		Evidence: []routecontract.Evidence{
			evidence,
		},
		Limitations: limitations,
	}
}

func endpointLimitations(
	input Input,
) []routecontract.Limitation {
	role := string(input.Candidates.Role)
	limitations := []routecontract.Limitation{
		{
			Code:    "probable_endpoint_only",
			Message: "The airport endpoint is inferred and is not filed or operational flight-plan data.",
			Scope:   role,
		},
	}

	if input.Candidates.Role ==
		routecontract.EndpointRoleDestination {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code:    "destination_not_planned_destination",
				Message: "The selected destination reflects persisted trajectory evidence and may not be the planned destination.",
				Scope:   role,
			},
		)
	}
	if input.CoverageGapCount > 0 {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code: "trajectory_coverage_gaps",
				Message: fmt.Sprintf(
					"The trajectory contains %d coverage gap(s), which may weaken endpoint inference.",
					input.CoverageGapCount,
				),
				Scope: role,
			},
		)
	}
	if input.Candidates.TruncatedCandidateCount > 0 {
		limitations = append(
			limitations,
			routecontract.Limitation{
				Code: "airport_candidates_truncated",
				Message: fmt.Sprintf(
					"%d geographically eligible airport candidate(s) were omitted by the configured result limit.",
					input.Candidates.
						TruncatedCandidateCount,
				),
				Scope: role,
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

func evidenceAttributes(
	input Input,
	candidate airportresolver.Candidate,
	scoreGap float64,
) []routecontract.EvidenceAttribute {
	attributes := []routecontract.EvidenceAttribute{
		{
			Key: "candidate_count",
			Value: strconv.Itoa(
				len(input.Candidates.Candidates),
			),
		},
		{
			Key:   "candidate_score_gap",
			Value: formatFloat(scoreGap),
		},
		{
			Key:   "catalog_fingerprint",
			Value: input.Candidates.CatalogFingerprint,
		},
		{
			Key:   "catalog_version",
			Value: input.Candidates.CatalogVersion,
		},
		{
			Key:   "distance_km",
			Value: formatFloat(candidate.DistanceKM),
		},
		{
			Key: "filtered_by_radius_count",
			Value: strconv.Itoa(
				input.Candidates.
					FilteredByRadiusCount,
			),
		},
		{
			Key: "maximum_distance_km",
			Value: formatFloat(
				input.Candidates.
					MaximumDistanceKM,
			),
		},
		{
			Key: "point_count",
			Value: strconv.Itoa(
				input.SegmentPointCount,
			),
		},
		{
			Key: "proximity_score",
			Value: formatFloat(
				candidate.ProximityScore,
			),
		},
		{
			Key:   "rank",
			Value: strconv.Itoa(candidate.Rank),
		},
		{
			Key:   "segment_status",
			Value: string(input.SegmentStatus),
		},
		{
			Key: "trajectory_quality",
			Value: formatFloat(
				input.TrajectoryQuality,
			),
		},
		{
			Key: "truncated_candidate_count",
			Value: strconv.Itoa(
				input.Candidates.
					TruncatedCandidateCount,
			),
		},
	}
	sort.SliceStable(
		attributes,
		func(left int, right int) bool {
			return attributes[left].Key <
				attributes[right].Key
		},
	)

	return attributes
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(
		value,
		'f',
		6,
		64,
	)
}

func inputFingerprint(
	input Input,
	minimumSelectionScore float64,
	minimumCandidateScoreGap float64,
) (string, error) {
	payload := struct {
		CandidateInputFingerprint string  `json:"candidate_input_fingerprint"`
		ObservedAt                string  `json:"observed_at"`
		TrajectoryQuality         float64 `json:"trajectory_quality"`
		SegmentStatus             string  `json:"segment_status"`
		SegmentPointCount         int     `json:"segment_point_count"`
		CoverageGapCount          int     `json:"coverage_gap_count"`
		MinimumSelectionScore     float64 `json:"minimum_selection_score"`
		MinimumCandidateScoreGap  float64 `json:"minimum_candidate_score_gap"`
	}{
		CandidateInputFingerprint: input.Candidates.InputFingerprint,
		ObservedAt: input.ObservedAt.Format(
			"2006-01-02T15:04:05.999999999Z07:00",
		),
		TrajectoryQuality:        input.TrajectoryQuality,
		SegmentStatus:            string(input.SegmentStatus),
		SegmentPointCount:        input.SegmentPointCount,
		CoverageGapCount:         input.CoverageGapCount,
		MinimumSelectionScore:    minimumSelectionScore,
		MinimumCandidateScoreGap: minimumCandidateScoreGap,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
