package server

import "testing"

func TestNewWeatherContextPostgresReaderRejectsNilPool(t *testing.T) {
	reader, err := NewWeatherContextPostgresReader(nil, nil)
	if err == nil {
		t.Fatalf("expected nil PostgreSQL pool error")
	}
	if reader != nil {
		t.Fatalf("reader must be nil when composition fails")
	}
}
