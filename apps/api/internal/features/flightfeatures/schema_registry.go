package flightfeatures

// SchemaCompatibility describes the relationship between two registered
// feature-schema versions.
type SchemaCompatibility string

const (
	SchemaCompatibilityIdentical   SchemaCompatibility = "identical"
	SchemaCompatibilityUnsupported SchemaCompatibility = "unsupported"
)

// SupportedSchemaVersions returns a defensive copy of every schema version
// that can be resolved by this package.
func SupportedSchemaVersions() []SchemaVersion {
	return []SchemaVersion{SchemaVersionV1}
}

// SchemaForVersion resolves an immutable defensive copy of one registered
// feature schema.
func SchemaForVersion(version SchemaVersion) (Schema, bool) {
	switch version {
	case SchemaVersionV1:
		return Schema{
			Version: version,
			Definitions: append(
				[]FeatureDefinition(nil),
				currentDefinitions...,
			),
		}, true
	default:
		return Schema{}, false
	}
}

// DefinitionByNameForVersion resolves one feature definition inside an
// explicitly selected schema version.
func DefinitionByNameForVersion(
	version SchemaVersion,
	name string,
) (FeatureDefinition, bool) {
	schema, found := SchemaForVersion(version)
	if !found {
		return FeatureDefinition{}, false
	}
	for _, definition := range schema.Definitions {
		if definition.Name == name {
			return definition, true
		}
	}
	return FeatureDefinition{}, false
}

// CompatibilityBetween reports whether two registered schema versions are
// identical. Additional compatibility states can be introduced when a second
// schema version exists and concrete migration semantics are known.
func CompatibilityBetween(
	left SchemaVersion,
	right SchemaVersion,
) SchemaCompatibility {
	if !left.IsValid() || !right.IsValid() {
		return SchemaCompatibilityUnsupported
	}
	if left == right {
		return SchemaCompatibilityIdentical
	}
	return SchemaCompatibilityUnsupported
}

func (value SchemaVersion) IsValid() bool {
	_, found := SchemaForVersion(value)
	return found
}

func (value AvailabilityStatus) IsValid() bool {
	switch value {
	case AvailabilityStatusAvailable,
		AvailabilityStatusPartial,
		AvailabilityStatusUnavailable:
		return true
	default:
		return false
	}
}

func (value ValidationStatus) IsValid() bool {
	switch value {
	case ValidationStatusUnvalidated,
		ValidationStatusValid,
		ValidationStatusLimited,
		ValidationStatusInvalid:
		return true
	default:
		return false
	}
}

func (value ValidationAuditState) IsValid() bool {
	switch value {
	case ValidationAuditStateComplete,
		ValidationAuditStateLegacyUnavailable:
		return true
	default:
		return false
	}
}

func (value ValidationIssueSeverity) IsValid() bool {
	switch value {
	case ValidationIssueSeverityWarning,
		ValidationIssueSeverityError:
		return true
	default:
		return false
	}
}

func (value FeatureGroup) IsValid() bool {
	switch value {
	case FeatureGroupTemporal,
		FeatureGroupGeographical,
		FeatureGroupOperational,
		FeatureGroupTrajectory,
		FeatureGroupAircraft:
		return true
	default:
		return false
	}
}

func (value FeatureValueType) IsValid() bool {
	switch value {
	case FeatureValueTypeBoolean,
		FeatureValueTypeFloat64,
		FeatureValueTypeInteger,
		FeatureValueTypeString:
		return true
	default:
		return false
	}
}

func (value AircraftEnrichmentMode) IsValid() bool {
	switch value {
	case AircraftEnrichmentModeEnabled,
		AircraftEnrichmentModeDisabled:
		return true
	default:
		return false
	}
}
