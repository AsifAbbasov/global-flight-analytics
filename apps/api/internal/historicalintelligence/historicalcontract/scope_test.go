package historicalcontract

import "testing"

func TestScopeEqualUsesCompleteDomainIdentity(
	t *testing.T,
) {
	left := Scope{
		Type:                ScopeTypeRoute,
		OriginICAOCode:      "UBBB",
		DestinationICAOCode: "LTFM",
	}
	right := left

	if !left.Equal(right) {
		t.Fatal("identical scopes must be equal")
	}

	right.DestinationICAOCode = "LTAI"
	if left.Equal(right) {
		t.Fatal("different route destinations must not be equal")
	}
}
