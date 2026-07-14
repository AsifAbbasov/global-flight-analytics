package historicalwindow

import (
	"regexp"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestPlanFingerprintIsDeterministic(
	t *testing.T,
) {
	request := Request{
		StartTime: time.Date(
			2026,
			time.July,
			1,
			10,
			15,
			0,
			0,
			time.UTC,
		),
		EndTime: time.Date(
			2026,
			time.July,
			1,
			14,
			30,
			0,
			0,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			1,
			13,
			40,
			0,
			0,
			time.UTC,
		),
		Granularity: historicalcontract.
			GranularityHour,
	}

	first := mustBuild(t, request)
	second := mustBuild(t, request)

	pattern := regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
	if first.Fingerprint !=
		second.Fingerprint ||
		!pattern.MatchString(
			first.Fingerprint,
		) {
		t.Fatalf(
			"unexpected fingerprints: %q %q",
			first.Fingerprint,
			second.Fingerprint,
		)
	}
}

func TestPlanFingerprintChangesWithCanonicalInput(
	t *testing.T,
) {
	base := Request{
		StartTime: time.Date(
			2026,
			time.July,
			1,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		EndTime: time.Date(
			2026,
			time.July,
			1,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			1,
			13,
			0,
			0,
			0,
			time.UTC,
		),
		Granularity: historicalcontract.
			GranularityHour,
	}

	first := mustBuild(t, base)

	changed := base
	changed.EndTime = changed.EndTime.Add(
		time.Hour,
	)
	second := mustBuild(t, changed)

	if first.Fingerprint ==
		second.Fingerprint {
		t.Fatal(
			"fingerprint did not change with end time",
		)
	}

	changed = base
	changed.Granularity =
		historicalcontract.GranularityCustom
	third := mustBuild(t, changed)

	if first.Fingerprint ==
		third.Fingerprint {
		t.Fatal(
			"fingerprint did not change with granularity",
		)
	}
}

func TestBucketKeysAreStableAndUnique(
	t *testing.T,
) {
	request := Request{
		StartTime: time.Date(
			2026,
			time.July,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		EndTime: time.Date(
			2026,
			time.July,
			1,
			3,
			0,
			0,
			0,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			1,
			4,
			0,
			0,
			0,
			time.UTC,
		),
		Granularity: historicalcontract.
			GranularityHour,
	}

	first := mustBuild(t, request)
	second := mustBuild(t, request)

	seen := make(map[string]struct{})
	for index, bucket := range first.Buckets {
		if bucket.Key !=
			second.Buckets[index].Key {
			t.Fatalf(
				"bucket key changed: %q %q",
				bucket.Key,
				second.Buckets[index].Key,
			)
		}
		if _, exists := seen[bucket.Key]; exists {
			t.Fatalf(
				"duplicate bucket key: %s",
				bucket.Key,
			)
		}
		seen[bucket.Key] = struct{}{}
	}
}
