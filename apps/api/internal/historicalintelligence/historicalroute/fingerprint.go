package historicalroute

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func routeFingerprint(
	request Request,
	items []routeEvidence,
	origin string,
	destination string,
) string {
	records := []string{
		Version,
		string(request.MetricName),
		origin,
		destination,
		request.Plan.Fingerprint,
		request.Plan.AsOfTime.UTC().Format(time.RFC3339Nano),
		fmt.Sprintf(
			"route_limit_reached|%t",
			request.Snapshot.RouteLimitReached,
		),
		fmt.Sprintf(
			"route_byte_limit_reached|%t",
			request.Snapshot.RouteByteLimitReached,
		),
	}

	if request.Snapshot.RouteLimitReached ||
		request.Snapshot.RouteByteLimitReached {
		records = append(
			records,
			fmt.Sprintf(
				"route_matched_count|%d",
				request.Snapshot.RouteMatchedCount,
			),
			fmt.Sprintf(
				"route_payload_bytes|%d|%d",
				request.Snapshot.RoutePayloadBytes,
				request.Snapshot.RouteTotalPayloadBytes,
			),
		)
	}

	for _, item := range items {
		records = append(records, fmt.Sprintf(
			"route|%s|%s|%s|%s|%s|%s|%s|%d|%s|%s",
			item.record.ID,
			item.record.TrajectoryID,
			item.record.AsOfTime.UTC().Format(time.RFC3339Nano),
			item.record.StoredAt.UTC().Format(time.RFC3339Nano),
			item.record.Status,
			item.record.ConfidenceLevel,
			item.record.InputFingerprint,
			item.record.ValidationWarningCount,
			item.record.PayloadDigest(),
			validatedRouteResultDigest(item),
		))
	}

	sort.Strings(records)
	sum := sha256.Sum256([]byte(strings.Join(records, "\n")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validatedRouteResultDigest(item routeEvidence) string {
	encoded, err := json.Marshal(item.result)
	if err != nil {
		return "sha256:unavailable"
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}
