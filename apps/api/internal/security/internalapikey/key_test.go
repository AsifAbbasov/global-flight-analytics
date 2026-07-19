package internalapikey

import (
	"errors"
	"strings"
	"testing"
)

func TestParseDigestHexRoundTrip(
	t *testing.T,
) {
	key := strings.Repeat(
		"stage14-key-",
		4,
	)
	digest := DigestCandidate(key)

	parsed, err := ParseDigestHex(
		digest.Hex(),
	)
	if err != nil {
		t.Fatalf(
			"parse digest: %v",
			err,
		)
	}
	if parsed != digest {
		t.Fatalf(
			"parsed digest = %x, want %x",
			parsed,
			digest,
		)
	}
}

func TestParseDigestHexRejectsInvalidValues(
	t *testing.T,
) {
	tests := []struct {
		name   string
		value  string
		target error
	}{
		{
			name:   "missing",
			target: ErrDigestRequired,
		},
		{
			name:   "short",
			value:  strings.Repeat("a", 63),
			target: ErrDigestLength,
		},
		{
			name:   "non hexadecimal",
			value:  strings.Repeat("z", 64),
			target: ErrDigestFormat,
		},
		{
			name:   "surrounding whitespace",
			value:  " " + strings.Repeat("a", 64),
			target: ErrDigestFormat,
		},
		{
			name:   "zero digest",
			value:  strings.Repeat("0", 64),
			target: ErrDigestZero,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				_, err := ParseDigestHex(
					test.value,
				)
				if !errors.Is(
					err,
					test.target,
				) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						test.target,
					)
				}
			},
		)
	}
}

func TestDigestMatchesOnlyCorrectBoundedCandidate(
	t *testing.T,
) {
	key := strings.Repeat(
		"correct-key-",
		4,
	)
	digest := DigestCandidate(key)

	if !digest.MatchesCandidate(key) {
		t.Fatal(
			"correct candidate did not match",
		)
	}
	if digest.MatchesCandidate(
		strings.Repeat(
			"wrong-key-",
			4,
		),
	) {
		t.Fatal(
			"incorrect candidate matched",
		)
	}
	if digest.MatchesCandidate("short") {
		t.Fatal(
			"short candidate matched",
		)
	}
	if digest.MatchesCandidate(
		strings.Repeat(
			"x",
			MaximumCandidateLength+1,
		),
	) {
		t.Fatal(
			"oversized candidate matched",
		)
	}
}
