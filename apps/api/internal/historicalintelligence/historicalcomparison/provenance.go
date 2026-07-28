package historicalcomparison

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func comparisonProvenance(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) historicalcontract.Provenance {
	return historicalcontract.Provenance{
		BuilderVersion: strings.Join(
			[]string{
				Version,
				strings.TrimSpace(
					current.Provenance.BuilderVersion,
				),
				strings.TrimSpace(
					previous.Provenance.BuilderVersion,
				),
			},
			"+",
		),
		InputFingerprint: comparisonFingerprint(
			current,
			previous,
		),
		SourceNames: mergeSourceNames(
			current.Provenance.SourceNames,
			previous.Provenance.SourceNames,
		),
		LatestSourceUpdatedAt: laterTime(
			current.Provenance.LatestSourceUpdatedAt,
			previous.Provenance.LatestSourceUpdatedAt,
		),
	}
}

func comparisonFingerprint(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) string {
	records := []string{Version}
	records = appendResultIdentity(
		records,
		"current",
		current,
	)
	records = appendResultIdentity(
		records,
		"previous",
		previous,
	)

	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func appendResultIdentity(
	records []string,
	prefix string,
	result historicalcontract.Result,
) []string {
	records = append(
		records,
		prefix+".schema="+string(result.SchemaVersion),
		prefix+".status="+string(result.Status),
		prefix+".metric_name="+string(result.Metric.Name),
		prefix+".metric_unit="+result.Metric.Unit,
		prefix+".aggregation="+string(
			result.Metric.Aggregation,
		),
		prefix+".scope="+scopeIdentity(result.Scope),
		prefix+".granularity="+string(
			result.Granularity,
		),
		prefix+".window_start="+formatTime(
			result.Window.StartTime,
		),
		prefix+".window_end="+formatTime(
			result.Window.EndTime,
		),
		prefix+".as_of="+formatTime(
			result.Window.AsOfTime,
		),
		prefix+".summary_point_count="+
			strconv.Itoa(result.Summary.PointCount),
		prefix+".summary_total="+formatFloat(
			result.Summary.Total,
		),
		prefix+".summary_minimum="+formatFloat(
			result.Summary.Minimum,
		),
		prefix+".summary_maximum="+formatFloat(
			result.Summary.Maximum,
		),
		prefix+".summary_average="+formatFloat(
			result.Summary.Average,
		),
		prefix+".summary_median="+formatFloat(
			result.Summary.Median,
		),
		prefix+".confidence_score="+formatFloat(
			result.Confidence.Score,
		),
		prefix+".confidence_samples="+
			strconv.Itoa(result.Confidence.SampleCount),
		prefix+".builder="+strings.TrimSpace(
			result.Provenance.BuilderVersion,
		),
		prefix+".input="+strings.TrimSpace(
			result.Provenance.InputFingerprint,
		),
		prefix+".latest_source="+formatTime(
			result.Provenance.LatestSourceUpdatedAt,
		),
		prefix+".generated_at="+formatTime(
			result.GeneratedAt,
		),
	)

	for index, point := range result.Points {
		pointPrefix := prefix + ".point." +
			strconv.Itoa(index)
		records = append(
			records,
			pointPrefix+".status="+string(point.Status),
			pointPrefix+".start="+formatTime(
				point.StartTime,
			),
			pointPrefix+".end="+formatTime(
				point.EndTime,
			),
			pointPrefix+".value="+formatFloat(
				point.Value,
			),
			pointPrefix+".samples="+
				strconv.Itoa(point.SampleCount),
			pointPrefix+".coverage="+formatFloat(
				point.CoverageRatio,
			),
		)
	}

	limitations := append(
		[]historicalcontract.Limitation(nil),
		result.Limitations...,
	)
	sort.SliceStable(
		limitations,
		func(left int, right int) bool {
			if limitations[left].Scope !=
				limitations[right].Scope {
				return limitations[left].Scope <
					limitations[right].Scope
			}
			return limitations[left].Code <
				limitations[right].Code
		},
	)
	for index, limitation := range limitations {
		limitationPrefix := prefix +
			".limitation." + strconv.Itoa(index)
		records = append(
			records,
			limitationPrefix+".scope="+
				limitation.Scope,
			limitationPrefix+".code="+
				limitation.Code,
			limitationPrefix+".message="+
				limitation.Message,
		)
	}

	sources := append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)
	sort.Strings(sources)
	for index, source := range sources {
		records = append(
			records,
			prefix+".source."+strconv.Itoa(index)+
				"="+source,
		)
	}
	return records
}

func scopeIdentity(
	scope historicalcontract.Scope,
) string {
	return strings.Join(
		[]string{
			string(scope.Type),
			scope.RegionCode,
			scope.AirportICAOCode,
			scope.OriginICAOCode,
			scope.DestinationICAOCode,
		},
		"|",
	)
}

func mergeSourceNames(
	groups ...[]string,
) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, group := range groups {
		for _, sourceName := range group {
			normalized := strings.TrimSpace(
				sourceName,
			)
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, normalized)
		}
	}

	sort.Strings(result)
	return result
}

func laterTime(
	left time.Time,
	right time.Time,
) time.Time {
	left = left.UTC()
	right = right.UTC()
	if right.After(left) {
		return right
	}
	return left
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(
		value,
		'g',
		-1,
		64,
	)
}
