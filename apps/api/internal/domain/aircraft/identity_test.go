package aircraft

import "testing"

func TestNormalizeICAO24(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
		canonical bool
	}{
		{name: "lowercase and spaces", input: " abc123 ", want: "ABC123", wantValid: true},
		{name: "already canonical", input: "ABC123", want: "ABC123", wantValid: true, canonical: true},
		{name: "invalid length", input: "ABC12", want: "ABC12", wantValid: false},
		{name: "invalid hexadecimal", input: "ABC12Z", want: "ABC12Z", wantValid: false},
		{name: "empty", input: "", want: "", wantValid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, valid := NormalizeICAO24(test.input)
			if got != test.want || valid != test.wantValid {
				t.Fatalf("NormalizeICAO24(%q) = (%q, %v), want (%q, %v)", test.input, got, valid, test.want, test.wantValid)
			}
			if IsCanonicalICAO24(test.input) != test.canonical {
				t.Fatalf("IsCanonicalICAO24(%q) = %v, want %v", test.input, IsCanonicalICAO24(test.input), test.canonical)
			}
		})
	}
}
