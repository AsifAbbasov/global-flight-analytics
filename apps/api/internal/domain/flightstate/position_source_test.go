package flightstate

import "testing"

func TestNormalizePositionSourceAcceptsCanonicalValues(t *testing.T) {
	values := []PositionSource{
		PositionSourceUnknown,
		PositionSourceADSB,
		PositionSourceADSR,
		PositionSourceTISB,
		PositionSourceADSC,
		PositionSourceASTERIX,
		PositionSourceMLAT,
		PositionSourceFLARM,
	}

	for _, value := range values {
		normalized, err := NormalizePositionSource(value)
		if err != nil {
			t.Fatalf("normalize %q: %v", value, err)
		}
		if normalized != value {
			t.Fatalf("normalize %q = %q", value, normalized)
		}
	}
}

func TestNormalizePositionSourceRejectsNonCanonicalReadsbType(t *testing.T) {
	if _, err := NormalizePositionSource("adsb_icao"); err == nil {
		t.Fatal("raw readsb type must be normalized at the integration boundary")
	}
}
