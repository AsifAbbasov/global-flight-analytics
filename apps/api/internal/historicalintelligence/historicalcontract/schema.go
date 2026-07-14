package historicalcontract

type FieldGroup string

const (
	FieldGroupMetric     FieldGroup = "metric"
	FieldGroupScope      FieldGroup = "scope"
	FieldGroupWindow     FieldGroup = "window"
	FieldGroupSeries     FieldGroup = "series"
	FieldGroupSummary    FieldGroup = "summary"
	FieldGroupComparison FieldGroup = "comparison"
	FieldGroupConfidence FieldGroup = "confidence"
	FieldGroupProvenance FieldGroup = "provenance"
)

type FieldValueType string

const (
	FieldValueTypeBoolean FieldValueType = "boolean"
	FieldValueTypeFloat64 FieldValueType = "float64"
	FieldValueTypeInteger FieldValueType = "integer"
	FieldValueTypeObject  FieldValueType = "object"
	FieldValueTypeString  FieldValueType = "string"
	FieldValueTypeTime    FieldValueType = "time"
)

type FieldDefinition struct {
	Name        string
	Group       FieldGroup
	ValueType   FieldValueType
	Unit        string
	Required    bool
	Description string
}

type Schema struct {
	Version     SchemaVersion
	Definitions []FieldDefinition
}

var currentDefinitions = []FieldDefinition{
	{
		Name:        "metric.name",
		Group:       FieldGroupMetric,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Versioned Historical Intelligence metric name.",
	},
	{
		Name:        "metric.unit",
		Group:       FieldGroupMetric,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Display and interpretation unit for the metric value.",
	},
	{
		Name:        "metric.aggregation",
		Group:       FieldGroupMetric,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Aggregation semantics used to produce each historical point.",
	},
	{
		Name:        "scope.type",
		Group:       FieldGroupScope,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Global, region, airport, or route analytical scope.",
	},
	{
		Name:        "scope.region_code",
		Group:       FieldGroupScope,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Normalized region code for a region-scoped series.",
	},
	{
		Name:        "scope.airport_icao_code",
		Group:       FieldGroupScope,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Normalized ICAO airport code for an airport-scoped series.",
	},
	{
		Name:        "scope.origin_icao_code",
		Group:       FieldGroupScope,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Normalized origin ICAO code for a route-scoped series.",
	},
	{
		Name:        "scope.destination_icao_code",
		Group:       FieldGroupScope,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Normalized destination ICAO code for a route-scoped series.",
	},
	{
		Name:        "window.start_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "Inclusive UTC beginning of the requested historical range.",
	},
	{
		Name:        "window.end_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "Exclusive UTC end of the requested historical range.",
	},
	{
		Name:        "window.as_of_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC analytical cutoff beyond which evidence must not be used.",
	},
	{
		Name:        "granularity",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Hour, day, week, or custom historical bucket granularity.",
	},
	{
		Name:        "points",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeObject,
		Required:    true,
		Description: "Deterministically ordered Historical Intelligence points.",
	},
	{
		Name:        "points.start_time",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "Inclusive UTC start time of one historical bucket.",
	},
	{
		Name:        "points.end_time",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "Exclusive UTC end time of one historical bucket.",
	},
	{
		Name:        "points.status",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Unavailable, partial, or complete bucket coverage status.",
	},
	{
		Name:        "points.value",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Non-negative historical metric value for the bucket.",
	},
	{
		Name:        "points.sample_count",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeInteger,
		Unit:        "items",
		Required:    true,
		Description: "Number of source samples represented by the bucket.",
	},
	{
		Name:        "points.coverage_ratio",
		Group:       FieldGroupSeries,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Normalized evidence coverage for the bucket.",
	},
	{
		Name:        "summary.point_count",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeInteger,
		Unit:        "items",
		Required:    true,
		Description: "Number of available or partial points included in summary statistics.",
	},
	{
		Name:        "summary.total",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Sum of all available or partial point values.",
	},
	{
		Name:        "summary.minimum",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Minimum available or partial point value.",
	},
	{
		Name:        "summary.maximum",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Maximum available or partial point value.",
	},
	{
		Name:        "summary.average",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Arithmetic mean of available or partial point values.",
	},
	{
		Name:        "summary.median",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Required:    true,
		Description: "Median of available or partial point values.",
	},
	{
		Name:        "comparison",
		Group:       FieldGroupComparison,
		ValueType:   FieldValueTypeObject,
		Required:    false,
		Description: "Comparison against the immediately preceding equivalent period.",
	},
	{
		Name:        "comparison.absolute_change",
		Group:       FieldGroupComparison,
		ValueType:   FieldValueTypeFloat64,
		Required:    false,
		Description: "Current value minus previous-period value.",
	},
	{
		Name:        "comparison.percentage_change",
		Group:       FieldGroupComparison,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "percent",
		Required:    false,
		Description: "Percentage change when the previous value is non-zero.",
	},
	{
		Name:        "comparison.direction",
		Group:       FieldGroupComparison,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Unavailable, down, flat, or up trend direction.",
	},
	{
		Name:        "confidence.score",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Overall Historical Intelligence confidence score.",
	},
	{
		Name:        "confidence.level",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Confidence level derived from the normalized score.",
	},
	{
		Name:        "confidence.sample_count",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeInteger,
		Unit:        "items",
		Required:    true,
		Description: "Total source sample count represented by the series.",
	},
	{
		Name:        "provenance.builder_version",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Version of the historical builder that produced the result.",
	},
	{
		Name:        "provenance.input_fingerprint",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "SHA-256 fingerprint of canonical historical inputs.",
	},
	{
		Name:        "provenance.latest_source_updated_at",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "Latest UTC update time among source records used by the result.",
	},
	{
		Name:        "generated_at",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC time at which the Historical Intelligence result was generated.",
	},
}

func CurrentSchema() Schema {
	return Schema{
		Version: SchemaVersionV1,
		Definitions: append(
			[]FieldDefinition(nil),
			currentDefinitions...,
		),
	}
}

func DefinitionByName(
	name string,
) (FieldDefinition, bool) {
	for _, definition := range currentDefinitions {
		if definition.Name == name {
			return definition, true
		}
	}

	return FieldDefinition{}, false
}
