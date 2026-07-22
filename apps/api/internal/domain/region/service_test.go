package region

import "testing"

func TestServiceListReturnsIndependentSlice(t *testing.T) {
	service := NewService()
	first := service.List()
	if len(first) == 0 {
		t.Fatal("expected region catalog")
	}
	originalCode := first[0].Code
	first[0].Code = "mutated"
	second := service.List()
	if second[0].Code != originalCode {
		t.Fatalf("internal region catalog was mutated: %q", second[0].Code)
	}
}
