package observability

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MetricsPath = "/internal/metrics"

	metricNamespace = "global_flight_analytics"
)

var defaultDurationBuckets = []float64{
	0.005,
	0.01,
	0.025,
	0.05,
	0.1,
	0.25,
	0.5,
	1,
	2.5,
	5,
	10,
}

type BuildInfo struct {
	Version  string
	Revision string
}

type Collector interface {
	Name() string
	WritePrometheus(
		ctx context.Context,
		builder *strings.Builder,
	) error
}

type Registry struct {
	mu sync.RWMutex

	build BuildInfo

	httpRequests  map[httpMetricKey]uint64
	httpDurations map[httpMetricKey]*histogram
	httpInFlight  int64

	providerRequests  map[providerMetricKey]uint64
	providerDurations map[string]*histogram

	ingestionCycles              map[string]uint64
	ingestionDurations           map[string]*histogram
	ingestionConsecutiveFailures int64
	ingestionNextDelaySeconds    float64

	collectors            []Collector
	collectionErrors      map[string]uint64
	collectionLastSuccess map[string]bool
}

type httpMetricKey struct {
	method      string
	route       string
	statusClass string
}

type providerMetricKey struct {
	provider string
	outcome  string
}

type histogram struct {
	buckets []float64
	counts  []uint64
	count   uint64
	sum     float64
}

func NewRegistry(
	build BuildInfo,
) *Registry {
	return &Registry{
		build: BuildInfo{
			Version:  normalizedBuildLabel(build.Version),
			Revision: normalizedBuildLabel(build.Revision),
		},
		httpRequests:          make(map[httpMetricKey]uint64),
		httpDurations:         make(map[httpMetricKey]*histogram),
		providerRequests:      make(map[providerMetricKey]uint64),
		providerDurations:     make(map[string]*histogram),
		ingestionCycles:       make(map[string]uint64),
		ingestionDurations:    make(map[string]*histogram),
		collectionErrors:      make(map[string]uint64),
		collectionLastSuccess: make(map[string]bool),
	}
}

func (
	registry *Registry,
) RegisterCollector(
	collector Collector,
) error {
	if registry == nil {
		return fmt.Errorf("observability registry is required")
	}
	if collector == nil {
		return fmt.Errorf("observability collector is required")
	}

	name := normalizeCollectorName(collector.Name())
	if name == "unknown" {
		return fmt.Errorf("observability collector name is required")
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	for _, existing := range registry.collectors {
		if normalizeCollectorName(existing.Name()) == name {
			return fmt.Errorf(
				"observability collector %q is already registered",
				name,
			)
		}
	}

	registry.collectors = append(registry.collectors, collector)
	registry.collectionLastSuccess[name] = false
	return nil
}

func (
	registry *Registry,
) BeginHTTPRequest() {
	if registry == nil {
		return
	}

	registry.mu.Lock()
	registry.httpInFlight++
	registry.mu.Unlock()
}

func (
	registry *Registry,
) FinishHTTPRequest(
	method string,
	route string,
	statusCode int,
	duration time.Duration,
) {
	if registry == nil {
		return
	}

	key := httpMetricKey{
		method:      normalizeMethod(method),
		route:       normalizeRoute(route),
		statusClass: normalizeStatusClass(statusCode),
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.httpInFlight > 0 {
		registry.httpInFlight--
	}
	registry.httpRequests[key]++

	histogramValue := registry.httpDurations[key]
	if histogramValue == nil {
		histogramValue = newHistogram(defaultDurationBuckets)
		registry.httpDurations[key] = histogramValue
	}
	histogramValue.observe(duration.Seconds())
}

func (
	registry *Registry,
) ObserveProviderRequest(
	provider string,
	outcome string,
	duration time.Duration,
) {
	if registry == nil {
		return
	}

	normalizedProvider := normalizeProvider(provider)
	key := providerMetricKey{
		provider: normalizedProvider,
		outcome:  normalizeProviderOutcome(outcome),
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.providerRequests[key]++
	histogramValue := registry.providerDurations[normalizedProvider]
	if histogramValue == nil {
		histogramValue = newHistogram(defaultDurationBuckets)
		registry.providerDurations[normalizedProvider] = histogramValue
	}
	histogramValue.observe(duration.Seconds())
}

func (
	registry *Registry,
) ObserveIngestionCycle(
	result string,
	duration time.Duration,
	consecutiveFailures int,
	nextDelay time.Duration,
) {
	if registry == nil {
		return
	}

	normalizedResult := normalizeIngestionResult(result)

	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.ingestionCycles[normalizedResult]++
	histogramValue := registry.ingestionDurations[normalizedResult]
	if histogramValue == nil {
		histogramValue = newHistogram(defaultDurationBuckets)
		registry.ingestionDurations[normalizedResult] = histogramValue
	}
	histogramValue.observe(duration.Seconds())

	if consecutiveFailures < 0 {
		consecutiveFailures = 0
	}
	registry.ingestionConsecutiveFailures = int64(consecutiveFailures)
	if nextDelay < 0 {
		nextDelay = 0
	}
	registry.ingestionNextDelaySeconds = nextDelay.Seconds()
}

func (
	registry *Registry,
) Render(
	ctx context.Context,
) string {
	if registry == nil {
		return ""
	}
	if ctx == nil {
		return ""
	}

	registry.mu.RLock()
	build := registry.build
	httpInFlight := registry.httpInFlight
	httpRequests := cloneHTTPRequests(registry.httpRequests)
	httpDurations := cloneHTTPHistograms(registry.httpDurations)
	providerRequests := cloneProviderRequests(registry.providerRequests)
	providerDurations := cloneStringHistograms(registry.providerDurations)
	ingestionCycles := cloneStringCounters(registry.ingestionCycles)
	ingestionDurations := cloneStringHistograms(registry.ingestionDurations)
	ingestionConsecutiveFailures := registry.ingestionConsecutiveFailures
	ingestionNextDelaySeconds := registry.ingestionNextDelaySeconds
	collectors := append([]Collector(nil), registry.collectors...)
	registry.mu.RUnlock()

	var builder strings.Builder
	writeBuildInfo(&builder, build)
	writeHTTPMetrics(
		&builder,
		httpInFlight,
		httpRequests,
		httpDurations,
	)
	writeProviderMetrics(
		&builder,
		providerRequests,
		providerDurations,
	)
	writeIngestionMetrics(
		&builder,
		ingestionCycles,
		ingestionDurations,
		ingestionConsecutiveFailures,
		ingestionNextDelaySeconds,
	)

	writeCollectorHealthHeaders(&builder)

	sort.Slice(
		collectors,
		func(left int, right int) bool {
			return normalizeCollectorName(collectors[left].Name()) <
				normalizeCollectorName(collectors[right].Name())
		},
	)

	for _, collector := range collectors {
		name := normalizeCollectorName(collector.Name())
		var collectorBuilder strings.Builder
		err := collector.WritePrometheus(ctx, &collectorBuilder)

		registry.mu.Lock()
		if err != nil {
			registry.collectionErrors[name]++
			registry.collectionLastSuccess[name] = false
		} else {
			registry.collectionLastSuccess[name] = true
		}
		collectionErrors := registry.collectionErrors[name]
		collectionSuccess := registry.collectionLastSuccess[name]
		registry.mu.Unlock()

		if err == nil {
			builder.WriteString(collectorBuilder.String())
		}
		writeCollectorHealth(
			&builder,
			name,
			collectionErrors,
			collectionSuccess,
		)
	}

	return builder.String()
}

func newHistogram(
	buckets []float64,
) *histogram {
	return &histogram{
		buckets: append([]float64(nil), buckets...),
		counts:  make([]uint64, len(buckets)),
	}
}

func (
	histogramValue *histogram,
) observe(
	value float64,
) {
	if value < 0 {
		value = 0
	}

	histogramValue.count++
	histogramValue.sum += value
	for index, upperBound := range histogramValue.buckets {
		if value <= upperBound {
			histogramValue.counts[index]++
		}
	}
}

func (
	histogramValue *histogram,
) clone() *histogram {
	if histogramValue == nil {
		return newHistogram(defaultDurationBuckets)
	}

	return &histogram{
		buckets: append([]float64(nil), histogramValue.buckets...),
		counts:  append([]uint64(nil), histogramValue.counts...),
		count:   histogramValue.count,
		sum:     histogramValue.sum,
	}
}

func normalizedBuildLabel(
	value string,
) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

func normalizeMethod(
	value string,
) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return strings.ToUpper(strings.TrimSpace(value))
	default:
		return "OTHER"
	}
}

func normalizeRoute(
	value string,
) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") {
		return "unmatched"
	}
	if strings.ContainsAny(value, "?#") || len(value) > 160 {
		return "unmatched"
	}
	return value
}

func normalizeStatusClass(
	statusCode int,
) string {
	if statusCode < 100 || statusCode > 599 {
		return "unknown"
	}
	return fmt.Sprintf("%dxx", statusCode/100)
}

func normalizeProvider(
	value string,
) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open_meteo", "open-meteo":
		return "open_meteo"
	case "airplanes_live", "airplanes.live":
		return "airplanes_live"
	case "opensky", "open_sky":
		return "opensky"
	default:
		return "unknown"
	}
}

func normalizeProviderOutcome(
	value string,
) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "success",
		"rate_limited",
		"unauthorized",
		"client_error",
		"server_error",
		"timeout",
		"network_error",
		"invalid_response":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}

func normalizeIngestionResult(
	value string,
) string {
	if strings.EqualFold(strings.TrimSpace(value), "success") {
		return "success"
	}
	return "failed"
}

func normalizeCollectorName(
	value string,
) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}

	var builder strings.Builder
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z':
			builder.WriteRune(character)
		case character >= '0' && character <= '9':
			builder.WriteRune(character)
		case character == '_' || character == '-':
			builder.WriteRune('_')
		default:
			return "unknown"
		}
	}
	return builder.String()
}

func cloneHTTPRequests(
	source map[httpMetricKey]uint64,
) map[httpMetricKey]uint64 {
	result := make(map[httpMetricKey]uint64, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneHTTPHistograms(
	source map[httpMetricKey]*histogram,
) map[httpMetricKey]*histogram {
	result := make(map[httpMetricKey]*histogram, len(source))
	for key, value := range source {
		result[key] = value.clone()
	}
	return result
}

func cloneProviderRequests(
	source map[providerMetricKey]uint64,
) map[providerMetricKey]uint64 {
	result := make(map[providerMetricKey]uint64, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneStringCounters(
	source map[string]uint64,
) map[string]uint64 {
	result := make(map[string]uint64, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneStringHistograms(
	source map[string]*histogram,
) map[string]*histogram {
	result := make(map[string]*histogram, len(source))
	for key, value := range source {
		result[key] = value.clone()
	}
	return result
}
