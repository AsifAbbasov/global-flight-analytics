package historicalaggregate

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestScopeKeyNormalizesSupportedScopes(
	t *testing.T,
) {
	tests := []struct {
		name  string
		scope historicalcontract.Scope
		want  string
	}{
		{
			name: "global",
			scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			want: "global",
		},
		{
			name: "region",
			scope: historicalcontract.Scope{
				Type:       historicalcontract.ScopeTypeRegion,
				RegionCode: " az ",
			},
			want: "region:AZ",
		},
		{
			name: "airport",
			scope: historicalcontract.Scope{
				Type:            historicalcontract.ScopeTypeAirport,
				AirportICAOCode: " ubbb ",
			},
			want: "airport:UBBB",
		},
		{
			name: "route",
			scope: historicalcontract.Scope{
				Type:                historicalcontract.ScopeTypeRoute,
				OriginICAOCode:      " ubbb ",
				DestinationICAOCode: " ugtb ",
			},
			want: "route:UBBB:UGTB",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				got, err := scopeKey(test.scope)
				if err != nil {
					t.Fatalf(
						"build scope key: %v",
						err,
					)
				}
				if got != test.want {
					t.Fatalf(
						"scope key = %q, want %q",
						got,
						test.want,
					)
				}
			},
		)
	}
}

func TestNormalizeScopeRejectsIncompleteRoute(
	t *testing.T,
) {
	_, err := normalizeScope(
		historicalcontract.Scope{
			Type:           historicalcontract.ScopeTypeRoute,
			OriginICAOCode: "UBBB",
		},
	)
	if !errors.Is(err, ErrScopeInvalid) {
		t.Fatalf(
			"expected invalid scope error, got %v",
			err,
		)
	}
}

func TestRecordIDIsDeterministic(
	t *testing.T,
) {
	key := ResultKey{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		MetricName: historicalcontract.
			MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},
		Granularity: historicalcontract.
			GranularityHour,
		Window: historicalcontract.TimeWindow{
			StartTime: aggregateTestTime().
				Add(-time.Hour),
			EndTime:  aggregateTestTime(),
			AsOfTime: aggregateTestTime(),
		},
	}
	encoded, err := encodeResultKey(key)
	if err != nil {
		t.Fatalf("encode result key: %v", err)
	}

	fingerprint := "sha256:" +
		strings.Repeat("a", 64)
	left := makeRecordID(encoded, fingerprint)
	right := makeRecordID(encoded, fingerprint)

	if left != right {
		t.Fatal("expected deterministic record identifier")
	}
	if len(left) !=
		len(recordIDPrefix)+64 {
		t.Fatalf(
			"unexpected record identifier length: %d",
			len(left),
		)
	}
}

func aggregateTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
