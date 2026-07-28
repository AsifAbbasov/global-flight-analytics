package historicalreplay

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func replayInputFingerprint(
	plan historicalwindow.Plan,
	metricName historicalcontract.MetricName,
	scope historicalcontract.Scope,
	granularity historicalcontract.Granularity,
	datasetLimit int,
	maximumBucketCount int,
	maximumWindowCount int,
	generatedAt time.Time,
) string {
	records := []string{
		FingerprintVersion,
		plan.Version,
		plan.Fingerprint,
		plan.RequestedStartTime.UTC().
			Format(time.RFC3339Nano),
		plan.RequestedEndTime.UTC().
			Format(time.RFC3339Nano),
		plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		string(metricName),
		string(scope.Type),
		scope.RegionCode,
		scope.AirportICAOCode,
		scope.OriginICAOCode,
		scope.DestinationICAOCode,
		string(granularity),
		strconv.Itoa(datasetLimit),
		strconv.Itoa(maximumBucketCount),
		strconv.Itoa(maximumWindowCount),
		generatedAt.UTC().
			Format(time.RFC3339Nano),
	}
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}
