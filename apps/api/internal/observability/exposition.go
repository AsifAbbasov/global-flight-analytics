package observability

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func writeBuildInfo(
	builder *strings.Builder,
	build BuildInfo,
) {
	writeHelpAndType(
		builder,
		metricNamespace+"_build_info",
		"Build information for the running Global Flight Analytics process.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_build_info{version=%s,revision=%s} 1\n",
		metricNamespace,
		quoteLabel(build.Version),
		quoteLabel(build.Revision),
	)
}

func writeHTTPMetrics(
	builder *strings.Builder,
	inFlight int64,
	requests map[httpMetricKey]uint64,
	durations map[httpMetricKey]*histogram,
) {
	writeHelpAndType(
		builder,
		metricNamespace+"_http_requests_in_flight",
		"Current number of API requests being processed.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_http_requests_in_flight %d\n",
		metricNamespace,
		inFlight,
	)

	writeHelpAndType(
		builder,
		metricNamespace+"_http_requests_total",
		"Completed API requests partitioned by bounded method, route template and status class.",
		"counter",
	)

	keys := make([]httpMetricKey, 0, len(requests))
	for key := range requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left int, right int) bool {
		if keys[left].method != keys[right].method {
			return keys[left].method < keys[right].method
		}
		if keys[left].route != keys[right].route {
			return keys[left].route < keys[right].route
		}
		return keys[left].statusClass < keys[right].statusClass
	})

	for _, key := range keys {
		labels := []label{
			{name: "method", value: key.method},
			{name: "route", value: key.route},
			{name: "status_class", value: key.statusClass},
		}
		fmt.Fprintf(
			builder,
			"%s_http_requests_total%s %d\n",
			metricNamespace,
			formatLabels(labels),
			requests[key],
		)
	}

	writeHelpAndType(
		builder,
		metricNamespace+"_http_request_duration_seconds",
		"API request duration in seconds partitioned by bounded method, route template and status class.",
		"histogram",
	)
	for _, key := range keys {
		writeHistogram(
			builder,
			metricNamespace+"_http_request_duration_seconds",
			[]label{
				{name: "method", value: key.method},
				{name: "route", value: key.route},
				{name: "status_class", value: key.statusClass},
			},
			durations[key],
		)
	}
}

func writeProviderMetrics(
	builder *strings.Builder,
	requests map[providerMetricKey]uint64,
	durations map[string]*histogram,
) {
	writeHelpAndType(
		builder,
		metricNamespace+"_provider_requests_total",
		"External provider requests partitioned by bounded provider and outcome values.",
		"counter",
	)

	keys := make([]providerMetricKey, 0, len(requests))
	for key := range requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left int, right int) bool {
		if keys[left].provider != keys[right].provider {
			return keys[left].provider < keys[right].provider
		}
		return keys[left].outcome < keys[right].outcome
	})
	for _, key := range keys {
		fmt.Fprintf(
			builder,
			"%s_provider_requests_total%s %d\n",
			metricNamespace,
			formatLabels([]label{
				{name: "provider", value: key.provider},
				{name: "outcome", value: key.outcome},
			}),
			requests[key],
		)
	}

	writeHelpAndType(
		builder,
		metricNamespace+"_provider_request_duration_seconds",
		"External provider request duration in seconds partitioned by bounded provider value.",
		"histogram",
	)
	providers := make([]string, 0, len(durations))
	for provider := range durations {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	for _, provider := range providers {
		writeHistogram(
			builder,
			metricNamespace+"_provider_request_duration_seconds",
			[]label{{name: "provider", value: provider}},
			durations[provider],
		)
	}
}

func writeIngestionMetrics(
	builder *strings.Builder,
	cycles map[string]uint64,
	durations map[string]*histogram,
	consecutiveFailures int64,
	nextDelaySeconds float64,
) {
	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_cycles_total",
		"Traffic ingestion cycles partitioned by bounded result value.",
		"counter",
	)
	results := make([]string, 0, len(cycles))
	for result := range cycles {
		results = append(results, result)
	}
	sort.Strings(results)
	for _, result := range results {
		fmt.Fprintf(
			builder,
			"%s_ingestion_cycles_total%s %d\n",
			metricNamespace,
			formatLabels([]label{{name: "result", value: result}}),
			cycles[result],
		)
	}

	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_cycle_duration_seconds",
		"Traffic ingestion cycle duration in seconds partitioned by bounded result value.",
		"histogram",
	)
	for _, result := range results {
		writeHistogram(
			builder,
			metricNamespace+"_ingestion_cycle_duration_seconds",
			[]label{{name: "result", value: result}},
			durations[result],
		)
	}

	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_consecutive_failures",
		"Current number of consecutive failed traffic ingestion cycles.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_ingestion_consecutive_failures %d\n",
		metricNamespace,
		consecutiveFailures,
	)

	writeHelpAndType(
		builder,
		metricNamespace+"_ingestion_next_delay_seconds",
		"Delay in seconds before the next traffic ingestion cycle.",
		"gauge",
	)
	fmt.Fprintf(
		builder,
		"%s_ingestion_next_delay_seconds %s\n",
		metricNamespace,
		formatFloat(nextDelaySeconds),
	)
}

func writeCollectorHealthHeaders(
	builder *strings.Builder,
) {
	writeHelpAndType(
		builder,
		metricNamespace+"_collector_errors_total",
		"Observability collector failures partitioned by bounded collector name.",
		"counter",
	)
	writeHelpAndType(
		builder,
		metricNamespace+"_collector_last_scrape_success",
		"Whether the most recent scrape collection succeeded for the collector.",
		"gauge",
	)
}

func writeCollectorHealth(
	builder *strings.Builder,
	collector string,
	errorsTotal uint64,
	lastSuccess bool,
) {
	labels := formatLabels([]label{{name: "collector", value: collector}})
	fmt.Fprintf(
		builder,
		"%s_collector_errors_total%s %d\n",
		metricNamespace,
		labels,
		errorsTotal,
	)

	successValue := 0
	if lastSuccess {
		successValue = 1
	}
	fmt.Fprintf(
		builder,
		"%s_collector_last_scrape_success%s %d\n",
		metricNamespace,
		labels,
		successValue,
	)
}

func writeHistogram(
	builder *strings.Builder,
	name string,
	labels []label,
	value *histogram,
) {
	if value == nil {
		value = newHistogram(defaultDurationBuckets)
	}

	for index, upperBound := range value.buckets {
		bucketLabels := append(
			append([]label(nil), labels...),
			label{name: "le", value: formatFloat(upperBound)},
		)
		fmt.Fprintf(
			builder,
			"%s_bucket%s %d\n",
			name,
			formatLabels(bucketLabels),
			value.counts[index],
		)
	}
	infiniteLabels := append(
		append([]label(nil), labels...),
		label{name: "le", value: "+Inf"},
	)
	fmt.Fprintf(
		builder,
		"%s_bucket%s %d\n",
		name,
		formatLabels(infiniteLabels),
		value.count,
	)
	fmt.Fprintf(
		builder,
		"%s_sum%s %s\n",
		name,
		formatLabels(labels),
		formatFloat(value.sum),
	)
	fmt.Fprintf(
		builder,
		"%s_count%s %d\n",
		name,
		formatLabels(labels),
		value.count,
	)
}

type label struct {
	name  string
	value string
}

func formatLabels(
	labels []label,
) string {
	if len(labels) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteByte('{')
	for index, current := range labels {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(current.name)
		builder.WriteByte('=')
		builder.WriteString(quoteLabel(current.value))
	}
	builder.WriteByte('}')
	return builder.String()
}

func quoteLabel(
	value string,
) string {
	return strconv.Quote(value)
}

func formatFloat(
	value float64,
) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func writeHelpAndType(
	builder *strings.Builder,
	name string,
	help string,
	metricType string,
) {
	fmt.Fprintf(builder, "# HELP %s %s\n", name, help)
	fmt.Fprintf(builder, "# TYPE %s %s\n", name, metricType)
}
