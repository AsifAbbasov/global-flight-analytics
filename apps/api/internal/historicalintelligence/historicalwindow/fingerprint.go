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
		plan.RequestedStartTime.UTC().
			Format(time.RFC3339Nano),
		plan.RequestedEndTime.UTC().
			Format(time.RFC3339Nano),
		plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		string(plan.Granularity),
		fmt.Sprintf(
			"%d",
			plan.MaximumBucketCount,
		),
		fmt.Sprintf(
			"%t",
			plan.TruncatedByAsOfTime,
		),
	}

	if plan.EffectiveWindow == nil {
		parts = append(
			parts,
			"effective-window:none",
		)
	} else {
		parts = append(
			parts,
			plan.EffectiveWindow.StartTime.
				UTC().
				Format(time.RFC3339Nano),
			plan.EffectiveWindow.EndTime.
				UTC().
				Format(time.RFC3339Nano),
		)
	}

	for _, exclusion := range plan.Exclusions {
		parts = append(
			parts,
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
