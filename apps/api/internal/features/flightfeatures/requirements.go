package flightfeatures

// GroupRequirementCounts describes how many fields in a feature group are
// required for core completeness and how many are optional enrichment.
type GroupRequirementCounts struct {
	Required int
	Optional int
}

func (counts GroupRequirementCounts) Total() int {
	return counts.Required + counts.Optional
}

// CurrentGroupRequirementCounts returns a fresh requirement map derived from
// the current versioned feature schema.
func CurrentGroupRequirementCounts() map[FeatureGroup]GroupRequirementCounts {
	counts := make(map[FeatureGroup]GroupRequirementCounts)
	for _, definition := range CurrentSchema().Definitions {
		current := counts[definition.Group]
		if definition.Required {
			current.Required++
		} else {
			current.Optional++
		}
		counts[definition.Group] = current
	}
	return counts
}
