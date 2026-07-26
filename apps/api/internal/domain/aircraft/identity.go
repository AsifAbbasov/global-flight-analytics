package aircraft

import "strings"

const ICAO24Length = 6

// CanonicalICAO24 trims surrounding whitespace and normalizes hexadecimal
// letters to uppercase without asserting that the result is valid.
func CanonicalICAO24(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// NormalizeICAO24 returns the canonical identity and whether it satisfies the
// six-character uppercase hexadecimal ICAO24 contract.
func NormalizeICAO24(value string) (string, bool) {
	normalized := CanonicalICAO24(value)
	return normalized, validCanonicalICAO24(normalized)
}

// IsValidICAO24 accepts any value whose canonical form is a valid ICAO24.
func IsValidICAO24(value string) bool {
	_, valid := NormalizeICAO24(value)
	return valid
}

// IsCanonicalICAO24 requires the supplied value itself to already be in the
// canonical form used by persisted feature identities.
func IsCanonicalICAO24(value string) bool {
	normalized, valid := NormalizeICAO24(value)
	return valid && normalized == value
}

func validCanonicalICAO24(value string) bool {
	if len(value) != ICAO24Length {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character < '0' || character > '9') &&
			(character < 'A' || character > 'F') {
			return false
		}
	}
	return true
}
