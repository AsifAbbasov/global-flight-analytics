package projectionread

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const routeHistoryEvidenceFingerprintVersion = "projection-route-history-evidence-v1"

var routeHistoryInputFingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

type routeHistoryEvidence struct {
	EvidenceID       string `json:"evidence_id"`
	TrajectoryID     string `json:"trajectory_id"`
	RouteRecordID    string `json:"route_record_id"`
	InputFingerprint string `json:"input_fingerprint"`
	AsOfTimeUnixNano int64  `json:"as_of_time_unix_nano"`
}

type routeHistoryEvidenceBoundary struct {
	ObservationCount       int
	DistinctDayCount       int
	RecentObservationCount int
	WindowStart            time.Time
	RecentWindowStart      time.Time
	AsOfTime               time.Time
	LastObservedAt         time.Time
}

func decodeRouteHistoryEvidence(
	payload []byte,
	boundary routeHistoryEvidenceBoundary,
) ([]routeHistoryEvidence, error) {
	if boundary.ObservationCount < 1 || len(payload) == 0 {
		return nil, fmt.Errorf(
			"route-history evidence payload is required",
		)
	}

	var items []routeHistoryEvidence
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf(
			"decode route-history evidence payload: %w",
			err,
		)
	}
	if len(items) != boundary.ObservationCount {
		return nil, fmt.Errorf(
			"route-history evidence count = %d, expected %d",
			len(items),
			boundary.ObservationCount,
		)
	}

	boundary.WindowStart = boundary.WindowStart.UTC()
	boundary.RecentWindowStart = boundary.RecentWindowStart.UTC()
	boundary.AsOfTime = boundary.AsOfTime.UTC()
	boundary.LastObservedAt = boundary.LastObservedAt.UTC()
	seenEvidenceIDs := make(map[string]struct{}, len(items))
	seenTrajectoryIDs := make(map[string]struct{}, len(items))
	seenRouteRecordIDs := make(map[string]struct{}, len(items))
	distinctDays := make(map[string]struct{}, len(items))
	recentObservationCount := 0
	latestObservedAt := time.Time{}

	for index := range items {
		item := &items[index]
		item.EvidenceID = strings.TrimSpace(item.EvidenceID)
		item.TrajectoryID = strings.TrimSpace(item.TrajectoryID)
		item.RouteRecordID = strings.TrimSpace(item.RouteRecordID)
		item.InputFingerprint = strings.TrimSpace(
			item.InputFingerprint,
		)
		observedAt := time.Unix(
			0,
			item.AsOfTimeUnixNano,
		).UTC()

		if item.EvidenceID == "" ||
			item.TrajectoryID == "" ||
			item.RouteRecordID == "" ||
			!routeHistoryInputFingerprintPattern.MatchString(
				item.InputFingerprint,
			) ||
			item.AsOfTimeUnixNano == 0 ||
			observedAt.Before(boundary.WindowStart) ||
			observedAt.After(boundary.AsOfTime) {
			return nil, fmt.Errorf(
				"route-history evidence item %d is invalid",
				index,
			)
		}
		if _, exists := seenEvidenceIDs[item.EvidenceID]; exists {
			return nil, fmt.Errorf(
				"route-history evidence identifier is duplicated",
			)
		}
		if _, exists := seenTrajectoryIDs[item.TrajectoryID]; exists {
			return nil, fmt.Errorf(
				"route-history trajectory identifier is duplicated",
			)
		}
		if _, exists := seenRouteRecordIDs[item.RouteRecordID]; exists {
			return nil, fmt.Errorf(
				"route-history record identifier is duplicated",
			)
		}
		seenEvidenceIDs[item.EvidenceID] = struct{}{}
		seenTrajectoryIDs[item.TrajectoryID] = struct{}{}
		seenRouteRecordIDs[item.RouteRecordID] = struct{}{}
		distinctDays[observedAt.Format("2006-01-02")] = struct{}{}
		if !observedAt.Before(boundary.RecentWindowStart) {
			recentObservationCount++
		}
		if latestObservedAt.IsZero() || observedAt.After(latestObservedAt) {
			latestObservedAt = observedAt
		}
	}

	if len(distinctDays) != boundary.DistinctDayCount ||
		recentObservationCount != boundary.RecentObservationCount ||
		!latestObservedAt.Equal(boundary.LastObservedAt) {
		return nil, fmt.Errorf(
			"route-history evidence aggregate mirrors are invalid",
		)
	}

	sortRouteHistoryEvidence(items)
	return append([]routeHistoryEvidence(nil), items...), nil
}

func routeHistoryEvidenceFingerprint(
	items []routeHistoryEvidence,
) string {
	canonical := append([]routeHistoryEvidence(nil), items...)
	sortRouteHistoryEvidence(canonical)

	digest := sha256.New()
	writeRouteHistoryFingerprintValue(
		digest,
		routeHistoryEvidenceFingerprintVersion,
	)
	for _, item := range canonical {
		writeRouteHistoryFingerprintValue(digest, item.EvidenceID)
		writeRouteHistoryFingerprintValue(digest, item.TrajectoryID)
		writeRouteHistoryFingerprintValue(digest, item.RouteRecordID)
		writeRouteHistoryFingerprintValue(
			digest,
			item.InputFingerprint,
		)
		writeRouteHistoryFingerprintValue(
			digest,
			fmt.Sprintf("%d", item.AsOfTimeUnixNano),
		)
	}

	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

func sortRouteHistoryEvidence(
	items []routeHistoryEvidence,
) {
	sort.Slice(items, func(left int, right int) bool {
		if items[left].EvidenceID != items[right].EvidenceID {
			return items[left].EvidenceID < items[right].EvidenceID
		}
		if items[left].TrajectoryID != items[right].TrajectoryID {
			return items[left].TrajectoryID < items[right].TrajectoryID
		}
		if items[left].RouteRecordID != items[right].RouteRecordID {
			return items[left].RouteRecordID < items[right].RouteRecordID
		}
		if items[left].InputFingerprint !=
			items[right].InputFingerprint {
			return items[left].InputFingerprint <
				items[right].InputFingerprint
		}
		return items[left].AsOfTimeUnixNano <
			items[right].AsOfTimeUnixNano
	})
}

type routeHistoryFingerprintWriter interface {
	Write([]byte) (int, error)
}

func writeRouteHistoryFingerprintValue(
	digest routeHistoryFingerprintWriter,
	value string,
) {
	_, _ = fmt.Fprintf(
		digest,
		"%d:%s|",
		len(value),
		value,
	)
}
