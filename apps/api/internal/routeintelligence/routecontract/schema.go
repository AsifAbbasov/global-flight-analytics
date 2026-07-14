package routecontract

type FieldGroup string

const (
	FieldGroupIdentity    FieldGroup = "identity"
	FieldGroupWindow      FieldGroup = "window"
	FieldGroupOrigin      FieldGroup = "origin"
	FieldGroupDestination FieldGroup = "destination"
	FieldGroupSummary     FieldGroup = "summary"
	FieldGroupConfidence  FieldGroup = "confidence"
	FieldGroupProvenance  FieldGroup = "provenance"
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
		Name:        "identity.trajectory_id",
		Group:       FieldGroupIdentity,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Stable identifier of the trajectory used for route inference.",
	},
	{
		Name:        "identity.identity_key",
		Group:       FieldGroupIdentity,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Stable cross-batch flight identity key when available.",
	},
	{
		Name:        "identity.icao24",
		Group:       FieldGroupIdentity,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Normalized six-character aircraft ICAO24 address.",
	},
	{
		Name:        "identity.callsign",
		Group:       FieldGroupIdentity,
		ValueType:   FieldValueTypeString,
		Required:    false,
		Description: "Normalized aircraft callsign used by inference evidence.",
	},
	{
		Name:        "window.start_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC start time of the trajectory evidence window.",
	},
	{
		Name:        "window.end_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC end time of the trajectory evidence window.",
	},
	{
		Name:        "window.as_of_time",
		Group:       FieldGroupWindow,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC analytical cutoff beyond which evidence must not be used.",
	},
	{
		Name:        "origin",
		Group:       FieldGroupOrigin,
		ValueType:   FieldValueTypeObject,
		Required:    false,
		Description: "Probable origin airport and supporting evidence.",
	},
	{
		Name:        "origin.distance_km",
		Group:       FieldGroupOrigin,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "kilometres",
		Required:    false,
		Description: "Distance from origin evidence geometry to the airport reference.",
	},
	{
		Name:        "origin.confidence.score",
		Group:       FieldGroupOrigin,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "ratio",
		Required:    false,
		Description: "Normalized confidence score for the probable origin.",
	},
	{
		Name:        "destination",
		Group:       FieldGroupDestination,
		ValueType:   FieldValueTypeObject,
		Required:    false,
		Description: "Probable destination airport and supporting evidence.",
	},
	{
		Name:        "destination.distance_km",
		Group:       FieldGroupDestination,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "kilometres",
		Required:    false,
		Description: "Distance from destination evidence geometry to the airport reference.",
	},
	{
		Name:        "destination.confidence.score",
		Group:       FieldGroupDestination,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "ratio",
		Required:    false,
		Description: "Normalized confidence score for the probable destination.",
	},
	{
		Name:        "summary.great_circle_distance_km",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "kilometres",
		Required:    false,
		Description: "Great-circle distance between resolved origin and destination airports.",
	},
	{
		Name:        "summary.same_airport",
		Group:       FieldGroupSummary,
		ValueType:   FieldValueTypeBoolean,
		Required:    false,
		Description: "Whether origin and destination resolve to the same ICAO airport.",
	},
	{
		Name:        "confidence.score",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Overall normalized route confidence score.",
	},
	{
		Name:        "confidence.level",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Overall route confidence level derived from the score.",
	},
	{
		Name:        "confidence.evidence_count",
		Group:       FieldGroupConfidence,
		ValueType:   FieldValueTypeInteger,
		Unit:        "items",
		Required:    true,
		Description: "Total number of endpoint evidence records.",
	},
	{
		Name:        "provenance.resolver_version",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "Version of the resolver that produced the result.",
	},
	{
		Name:        "provenance.input_fingerprint",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeString,
		Required:    true,
		Description: "SHA-256 fingerprint of canonical route inference inputs.",
	},
	{
		Name:        "provenance.trajectory_updated_at",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC update time of the trajectory used by the resolver.",
	},
	{
		Name:        "generated_at",
		Group:       FieldGroupProvenance,
		ValueType:   FieldValueTypeTime,
		Required:    true,
		Description: "UTC time at which the route result was generated.",
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
