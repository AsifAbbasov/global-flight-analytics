package postgres

import "testing"

func TestNormalizeICAO24Lookup(t *testing.T) {
	result := normalizeICAO24Lookup("  abc123  ")

	if result != "ABC123" {
		t.Fatalf("expected ABC123, got %s", result)
	}
}

func TestNormalizeICAO24LookupEmptyValue(t *testing.T) {
	result := normalizeICAO24Lookup("   ")

	if result != "" {
		t.Fatalf("expected empty string, got %s", result)
	}
}
