package metrics

import (
	"errors"
	"math"
	"testing"
)

func TestActiveAircraftQueryScopesAreExplicit(t *testing.T) {
	global := NewGlobalActiveAircraftQueryScope()
	if global.Type != ActiveAircraftQueryScopeGlobal || global.IsBounded() {
		t.Fatalf("global scope = %#v", global)
	}

	bounded, err := NewBoundedActiveAircraftQueryScope(Bounds{
		MinLatitude:  38,
		MaxLatitude:  42,
		MinLongitude: 44,
		MaxLongitude: 51,
	})
	if err != nil {
		t.Fatalf("bounded scope error = %v", err)
	}
	if bounded.Type != ActiveAircraftQueryScopeBounds || !bounded.IsBounded() {
		t.Fatalf("bounded scope = %#v", bounded)
	}
}

func TestActiveAircraftQueryScopeRejectsInvalidEvidence(t *testing.T) {
	_, err := NewBoundedActiveAircraftQueryScope(Bounds{
		MinLatitude:  math.NaN(),
		MaxLatitude:  42,
		MinLongitude: 44,
		MaxLongitude: 51,
	})
	if !errors.Is(err, ErrActiveAircraftBoundsInvalid) {
		t.Fatalf("invalid bounds error = %v", err)
	}

	err = (ActiveAircraftQueryScope{Type: "unsupported"}).Validate()
	if !errors.Is(err, ErrActiveAircraftQueryScopeInvalid) {
		t.Fatalf("invalid scope error = %v", err)
	}
}
