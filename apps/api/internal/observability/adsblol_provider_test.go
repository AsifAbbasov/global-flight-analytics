package observability

import "testing"

func TestNormalizeProviderADSBLOL(t *testing.T) {
	if got := normalizeProvider("adsb.lol"); got != "adsb_lol" {
		t.Fatalf("normalized provider = %q", got)
	}
	if got := normalizeProvider("ADSB_LOL"); got != "adsb_lol" {
		t.Fatalf("normalized provider alias = %q", got)
	}
}
