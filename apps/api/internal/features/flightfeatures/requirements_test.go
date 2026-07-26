package flightfeatures

import "testing"

func TestCurrentGroupRequirementCountsRemainExplicit(t *testing.T) {
	counts := CurrentGroupRequirementCounts()

	wantRequired := map[FeatureGroup]int{
		FeatureGroupTemporal:     8,
		FeatureGroupGeographical: 11,
		FeatureGroupOperational:  11,
		FeatureGroupTrajectory:   16,
		FeatureGroupAircraft:     0,
	}
	wantOptional := map[FeatureGroup]int{
		FeatureGroupTemporal:     0,
		FeatureGroupGeographical: 0,
		FeatureGroupOperational:  0,
		FeatureGroupTrajectory:   0,
		FeatureGroupAircraft:     6,
	}

	for group, required := range wantRequired {
		got := counts[group]
		if got.Required != required || got.Optional != wantOptional[group] {
			t.Fatalf("group %q counts = %#v", group, got)
		}
		if got.Required > 0 && got.Optional > 0 {
			t.Fatalf("group %q mixes required and optional aggregate evidence", group)
		}
	}
}

func TestCurrentGroupFieldCountIsSchemaDerived(t *testing.T) {
	for group, counts := range CurrentGroupRequirementCounts() {
		if got := CurrentGroupFieldCount(group); got != counts.Total() {
			t.Fatalf("group %q field count = %d, want %d", group, got, counts.Total())
		}
	}
	if got := CurrentGroupFieldCount(FeatureGroupAircraft); got != 6 {
		t.Fatalf("aircraft field count = %d, want 6", got)
	}
}
