package region

import "testing"

func TestAzerbaijanRegionIsAvailable(t *testing.T) {
	selectedRegion, err := NewService().GetByCode("azerbaijan")
	if err != nil {
		t.Fatalf("expected Azerbaijan region, got %v", err)
	}

	if selectedRegion.Name != "Azerbaijan" {
		t.Fatalf("expected Azerbaijan name, got %q", selectedRegion.Name)
	}
	if selectedRegion.Bounds.MinLatitude != 38 ||
		selectedRegion.Bounds.MaxLatitude != 42 ||
		selectedRegion.Bounds.MinLongitude != 44.5 ||
		selectedRegion.Bounds.MaxLongitude != 51 {
		t.Fatalf(
			"unexpected Azerbaijan bounds: %#v",
			selectedRegion.Bounds,
		)
	}
}
