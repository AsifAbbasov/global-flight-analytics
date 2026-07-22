package reconciliation

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNormalizeLastErrorPreservesUTF8Boundary(t *testing.T) {
	value := strings.Repeat("a", maximumLastErrorLength-1) + "🙂"
	normalized := NormalizeLastError(value)
	if !utf8.ValidString(normalized) {
		t.Fatalf("normalized error is not valid UTF-8: %q", normalized)
	}
	if len(normalized) > maximumLastErrorLength {
		t.Fatalf("normalized byte length = %d, want <= %d", len(normalized), maximumLastErrorLength)
	}
	if strings.ContainsRune(normalized, utf8.RuneError) {
		t.Fatalf("normalization introduced a replacement rune: %q", normalized)
	}
}
