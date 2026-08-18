package providerpolicy

import "testing"

func TestADSBLOLUsesApplicationDefinedSafetyCap(t *testing.T) {
	policy, err := Get(ProviderADSBLOL)
	if err != nil {
		t.Fatal(err)
	}
	if policy.BudgetMode != BudgetModeFixedWindow {
		t.Fatalf("budget mode = %s", policy.BudgetMode)
	}
	if len(policy.RequestLimits) != 3 {
		t.Fatalf("request limits = %d", len(policy.RequestLimits))
	}

	minute := policy.RequestLimits[0]
	if minute.MaxRequests != 6 ||
		minute.Window != WindowMinute ||
		minute.Provenance != ProvenanceApplicationDefined {
		t.Fatalf("unexpected minute cap: %+v", minute)
	}

	day := policy.RequestLimits[2]
	if day.MaxRequests != 8640 ||
		day.Window != WindowDay ||
		day.Provenance != ProvenanceApplicationDefined {
		t.Fatalf("unexpected daily cap: %+v", day)
	}
}
