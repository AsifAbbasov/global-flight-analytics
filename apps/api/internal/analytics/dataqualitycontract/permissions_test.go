package dataqualitycontract

import (
	"errors"
	"testing"
)

func TestDeniedPermissionRequiresReason(t *testing.T) {
	_, err := DeniedPermission()
	if !errors.Is(err, ErrPermissionReasonRequired) {
		t.Fatalf("expected reason-required error, got %v", err)
	}
}

func TestPermissionCloneOwnsReasons(t *testing.T) {
	permission, err := DeniedPermission("sampling_density_below_threshold")
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	clone := permission.Clone()
	clone.Reasons[0] = "mutated"
	if permission.Reasons[0] != "sampling_density_below_threshold" {
		t.Fatal("expected cloned reason storage")
	}
}
