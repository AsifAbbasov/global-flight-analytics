package flightfeatures

const (
	TemporalRequiredFeatureFieldCount     = 8
	GeographicalRequiredFeatureFieldCount = 15
	OperationalRequiredFeatureFieldCount  = 11
	TrajectoryRequiredFeatureFieldCount   = 16
	AircraftOptionalFeatureFieldCount     = 6
)

// GroupRequirementCounts describes how many fields in a feature group are
// required for core completeness and how many are optional enrichment.
type GroupRequirementCounts struct {
	Required int
	Optional int
}

func (counts GroupRequirementCounts) Total() int {
	return counts.Required + counts.Optional
}

// GroupRequirementCountsForVersion returns a fresh requirement map derived
// from an explicitly selected registered feature schema.
func GroupRequirementCountsForVersion(
	version SchemaVersion,
) (map[FeatureGroup]GroupRequirementCounts, bool) {
	schema, found := SchemaForVersion(version)
	if !found {
		return nil, false
	}
	counts := make(map[FeatureGroup]GroupRequirementCounts)
	for _, definition := range schema.Definitions {
		current := counts[definition.Group]
		if definition.Required {
			current.Required++
		} else {
			current.Optional++
		}
		counts[definition.Group] = current
	}
	return counts, true
}

// CurrentGroupRequirementCounts returns a fresh requirement map derived from
// the current versioned feature schema.
func CurrentGroupRequirementCounts() map[FeatureGroup]GroupRequirementCounts {
	counts, found := GroupRequirementCountsForVersion(SchemaVersionV1)
	if !found {
		panic("current flight feature schema is not registered")
	}
	return counts
}

// GroupFieldCountForVersion returns one group field count for an explicitly
// selected registered schema version.
func GroupFieldCountForVersion(
	version SchemaVersion,
	group FeatureGroup,
) (int, bool) {
	counts, found := GroupRequirementCountsForVersion(version)
	if !found {
		return 0, false
	}
	return counts[group].Total(), true
}

// CurrentGroupFieldCount returns the total field count for one group directly
// from the current versioned feature schema.
func CurrentGroupFieldCount(group FeatureGroup) int {
	count, found := GroupFieldCountForVersion(SchemaVersionV1, group)
	if !found {
		panic("current flight feature schema is not registered")
	}
	return count
}
