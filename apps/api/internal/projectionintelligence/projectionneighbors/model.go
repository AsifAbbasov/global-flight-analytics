package projectionneighbors

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

const (
	Version            = "projection-historical-neighbor-selection-v1"
	FingerprintVersion = "projection-historical-neighbor-selection-fingerprint-v1"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusPartial     Status = "partial"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable,
		StatusPartial,
		StatusComplete:
		return true
	default:
		return false
	}
}

type RejectionCode string

const (
	RejectionSameTrajectory          RejectionCode = "same_trajectory"
	RejectionIdentifierMissing       RejectionCode = "candidate_identifier_missing"
	RejectionDuplicateCandidate      RejectionCode = "duplicate_candidate"
	RejectionNotHistorical           RejectionCode = "candidate_not_historical"
	RejectionTooOld                  RejectionCode = "candidate_too_old"
	RejectionInsufficientPoints      RejectionCode = "candidate_insufficient_points"
	RejectionContinuationUnavailable RejectionCode = "candidate_continuation_unavailable"
	RejectionAnchorTooDistant        RejectionCode = "candidate_anchor_too_distant"
	RejectionSimilarityUnavailable   RejectionCode = "candidate_similarity_unavailable"
	RejectionSimilarityBelowMinimum  RejectionCode = "candidate_similarity_below_minimum"
)

type Notice struct {
	Code    string
	Message string
}

type Rejection struct {
	TrajectoryID string
	Code         RejectionCode
	Message      string
}

type Neighbor struct {
	TrajectoryID string

	SimilarityScore            float64
	SimilarityLevel            historicalsimilarity.Level
	SimilarityInputFingerprint string

	AnchorPointIndex int
	AnchorObservedAt time.Time
	AnchorDistanceKM float64

	CandidateStartTime time.Time
	CandidateEndTime   time.Time
	CandidateAge       time.Duration

	PrefixPointCount       int
	ContinuationPointCount int
	ContinuationEndTime    time.Time
}

func (neighbor Neighbor) Clone() Neighbor {
	return neighbor
}

type Result struct {
	Version string
	Status  Status

	CurrentTrajectoryID          string
	AsOfTime                     time.Time
	RequiredContinuationDuration time.Duration

	InputCandidateCount     int
	CheckedCandidateCount   int
	QualifiedCandidateCount int
	RejectedCandidateCount  int

	SelectionLimit int
	Truncated      bool

	Neighbors   []Neighbor
	Rejections  []Rejection
	Limitations []Notice

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Neighbors = append(
		[]Neighbor(nil),
		result.Neighbors...,
	)
	cloned.Rejections = append(
		[]Rejection(nil),
		result.Rejections...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version {
		return fmt.Errorf(
			"neighbor selection version is invalid: %q",
			result.Version,
		)
	}
	if !result.Status.IsKnown() {
		return fmt.Errorf(
			"neighbor selection status is invalid: %q",
			result.Status,
		)
	}
	if strings.TrimSpace(
		result.CurrentTrajectoryID,
	) == "" {
		return fmt.Errorf(
			"current trajectory identifier is required",
		)
	}
	if result.AsOfTime.IsZero() ||
		result.RequiredContinuationDuration <= 0 {
		return fmt.Errorf(
			"as-of time and positive continuation duration are required",
		)
	}
	if result.InputCandidateCount < 0 ||
		result.CheckedCandidateCount < 0 ||
		result.QualifiedCandidateCount < 0 ||
		result.RejectedCandidateCount < 0 ||
		result.SelectionLimit < 1 {
		return fmt.Errorf(
			"neighbor selection counts are invalid",
		)
	}
	if result.CheckedCandidateCount >
		result.InputCandidateCount {
		return fmt.Errorf(
			"checked candidate count exceeds input candidate count",
		)
	}
	if result.QualifiedCandidateCount+
		result.RejectedCandidateCount !=
		result.CheckedCandidateCount {
		return fmt.Errorf(
			"qualified and rejected candidate counts do not match checked candidates",
		)
	}
	if len(result.Neighbors) >
		result.QualifiedCandidateCount {
		return fmt.Errorf(
			"selected neighbor count exceeds qualified candidate count",
		)
	}
	if result.RejectedCandidateCount !=
		len(result.Rejections) {
		return fmt.Errorf(
			"rejected candidate count does not match rejections",
		)
	}
	if len(result.Neighbors) >
		result.SelectionLimit {
		return fmt.Errorf(
			"selected neighbor count exceeds selection limit",
		)
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"neighbor selection input fingerprint is invalid",
		)
	}

	switch result.Status {
	case StatusUnavailable:
		if len(result.Neighbors) != 0 ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"unavailable selection requires no neighbors and at least one limitation",
			)
		}
	case StatusComplete:
		if len(result.Neighbors) !=
			result.SelectionLimit ||
			result.Truncated {
			return fmt.Errorf(
				"complete selection must fill the limit without truncation",
			)
		}
	case StatusPartial:
		if len(result.Neighbors) == 0 {
			return fmt.Errorf(
				"partial selection requires at least one neighbor",
			)
		}
	}

	seen := make(map[string]struct{})
	for index, neighbor := range result.Neighbors {
		if strings.TrimSpace(
			neighbor.TrajectoryID,
		) == "" ||
			neighbor.TrajectoryID ==
				result.CurrentTrajectoryID {
			return fmt.Errorf(
				"neighbor trajectory identifier is invalid",
			)
		}
		if _, exists := seen[neighbor.TrajectoryID]; exists {
			return fmt.Errorf(
				"duplicate selected neighbor: %s",
				neighbor.TrajectoryID,
			)
		}
		seen[neighbor.TrajectoryID] =
			struct{}{}

		if !unitInterval(
			neighbor.SimilarityScore,
		) ||
			neighbor.SimilarityLevel !=
				historicalsimilarity.
					LevelForScore(
						neighbor.
							SimilarityScore,
					) ||
			!fingerprintPattern.MatchString(
				neighbor.
					SimilarityInputFingerprint,
			) ||
			neighbor.AnchorPointIndex < 0 ||
			neighbor.AnchorObservedAt.IsZero() ||
			!finite(neighbor.AnchorDistanceKM) ||
			neighbor.AnchorDistanceKM < 0 ||
			neighbor.CandidateStartTime.IsZero() ||
			neighbor.CandidateEndTime.IsZero() ||
			neighbor.CandidateEndTime.Before(
				neighbor.CandidateStartTime,
			) ||
			neighbor.CandidateAge < 0 ||
			neighbor.PrefixPointCount < 2 ||
			neighbor.ContinuationPointCount < 1 ||
			neighbor.ContinuationEndTime.IsZero() ||
			neighbor.ContinuationEndTime.Before(
				neighbor.AnchorObservedAt.Add(
					result.
						RequiredContinuationDuration,
				),
			) {
			return fmt.Errorf(
				"selected neighbor is invalid: %s",
				neighbor.TrajectoryID,
			)
		}

		if index > 0 {
			previous := result.Neighbors[index-1]
			if previous.SimilarityScore <
				neighbor.SimilarityScore ||
				(previous.SimilarityScore ==
					neighbor.SimilarityScore &&
					previous.AnchorDistanceKM >
						neighbor.
							AnchorDistanceKM) ||
				(previous.SimilarityScore ==
					neighbor.SimilarityScore &&
					previous.AnchorDistanceKM ==
						neighbor.
							AnchorDistanceKM &&
					previous.TrajectoryID >
						neighbor.
							TrajectoryID) {
				return fmt.Errorf(
					"selected neighbors are not deterministically ordered",
				)
			}
		}
	}

	for _, rejection := range result.Rejections {
		if strings.TrimSpace(
			rejection.Message,
		) == "" ||
			rejection.Code == "" {
			return fmt.Errorf(
				"candidate rejection is invalid",
			)
		}
	}
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" {
			return fmt.Errorf(
				"neighbor selection limitation is invalid",
			)
		}
	}

	return nil
}

func normalizeNotices(
	items []Notice,
) []Notice {
	seen := make(map[string]Notice)
	for _, item := range items {
		key := strings.TrimSpace(
			item.Code,
		) + "\x00" +
			strings.TrimSpace(
				item.Message,
			)
		if key == "\x00" {
			continue
		}
		seen[key] = Notice{
			Code: strings.TrimSpace(
				item.Code,
			),
			Message: strings.TrimSpace(
				item.Message,
			),
		}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]Notice,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}
