package historicalwindow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func planFingerprint(
	plan Plan,
) string {
	parts := []string{
		FingerprintVersion,
		plan.Version,
		plan.RequestedStartTime.UTC().
			Format(time.RFC3339Nano),
		plan.RequestedEndTime.UTC().
			Format(time.RFC3339Nano),
		plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		string(plan.Granularity),
		fmt.Sprintf(
			"truncated:%t",
			plan.TruncatedByAsOfTime,
		),
	}

	parts = appendWindowFingerprint(
		parts,
		"effective",
		plan.EffectiveWindow,
	)
	parts = appendWindowFingerprint(
		parts,
		"previous",
		plan.PreviousWindow,
	)

	for _, bucket := range plan.Buckets {
		parts = append(
			parts,
			"bucket",
			bucket.Key,
			fmt.Sprintf("%d", bucket.Sequence),
			bucket.StartTime.UTC().
				Format(time.RFC3339Nano),
			bucket.EndTime.UTC().
				Format(time.RFC3339Nano),
		)
	}

	for _, exclusion := range plan.Exclusions {
		parts = append(
			parts,
			"exclusion",
			string(exclusion.Reason),
			exclusion.StartTime.UTC().
				Format(time.RFC3339Nano),
			exclusion.EndTime.UTC().
				Format(time.RFC3339Nano),
		)
	}

	sum := sha256.Sum256(
		[]byte(strings.Join(parts, "\x00")),
	)

	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func appendWindowFingerprint(
	parts []string,
	label string,
	window *historicalcontract.TimeWindow,
) []string {
	if window == nil {
		return append(parts, label+":none")
	}

	return append(
		parts,
		label,
		window.StartTime.UTC().
			Format(time.RFC3339Nano),
		window.EndTime.UTC().
			Format(time.RFC3339Nano),
		window.AsOfTime.UTC().
			Format(time.RFC3339Nano),
	)
}

func bucketKey(
	granularity historicalcontract.Granularity,
	startTime time.Time,
	endTime time.Time,
) string {
	canonical := strings.Join(
		[]string{
			Version,
			string(granularity),
			startTime.UTC().
				Format(time.RFC3339Nano),
			endTime.UTC().
				Format(time.RFC3339Nano),
		},
		"\x00",
	)

	sum := sha256.Sum256(
		[]byte(canonical),
	)

	return BucketKeyPrefix +
		hex.EncodeToString(sum[:])
}
